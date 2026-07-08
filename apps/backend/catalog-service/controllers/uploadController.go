package controllers

import (
	"catalog-service/models"
	"catalog-service/repository"
	"catalog-service/response"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// maxUploadSize borne la taille du fichier envoye : evite qu'un vendeur
// (ou un attaquant) sature le volume monte avec des fichiers enormes.
const maxUploadSize = 5 << 20 // 5 Mo

// allowedImageTypes mappe le type MIME reellement detecte (jamais celui
// declare par le client) vers l'extension de stockage.
var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// uploadDir renvoie le repertoire de stockage des photos, monte en volume
// persistant (voir infra/k8s/base/catalog-service/pvc.yaml).
func uploadDir() string {
	if d := os.Getenv("UPLOAD_DIR"); d != "" {
		return d
	}
	return "/data/uploads"
}

// UploadArticleImage recoit une photo pour une annonce existante (multipart,
// champ "image"). Deux garde-fous essentiels contre un fichier malveillant :
//   - le type reel du fichier est detecte via ses premiers octets
//     (http.DetectContentType), jamais via l'extension ou le Content-Type
//     declares par le client (tous deux falsifiables) ;
//   - le nom stocke est genere cote serveur (UUID), jamais celui envoye par
//     le client, ce qui elimine tout risque de path traversal.
//
// Le dossier de stockage est servi en pur statique par ailleurs (voir
// routes.go), sans possibilite d'execution.
func UploadArticleImage(c *gin.Context) {
	id := c.Param("id")
	var article models.Article
	if err := repository.DB.First(&article, "id = ?", id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Article introuvable")
		return
	}
	if !isAdmin(c) && article.SellerID != currentUserID(c) {
		response.Error(c, http.StatusForbidden, "Vous ne pouvez modifier que vos propres annonces")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Fichier image manquant ou trop volumineux (5 Mo max)")
		return
	}
	defer func() { _ = file.Close() }()

	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		response.Error(c, http.StatusBadRequest, "Fichier illisible")
		return
	}
	contentType := http.DetectContentType(head[:n])
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		response.Error(c, http.StatusBadRequest, "Format non supporte (jpeg, png, webp, gif uniquement)")
		return
	}

	dir := uploadDir()
	if err := os.MkdirAll(dir, 0o750); err != nil {
		response.Error(c, http.StatusInternalServerError, "Stockage indisponible")
		return
	}

	filename := uuid.New().String() + ext
	dest, err := os.OpenFile(filepath.Join(dir, filename), os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o640)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible d'enregistrer la photo")
		return
	}
	defer func() { _ = dest.Close() }()

	if _, err := dest.Write(head[:n]); err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible d'enregistrer la photo")
		return
	}
	if _, err := io.Copy(dest, file); err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible d'enregistrer la photo")
		return
	}

	article.ImageURL = "/uploads/" + filename
	if err := repository.DB.Model(&article).Update("image_url", article.ImageURL).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Impossible de mettre a jour l'annonce")
		return
	}

	c.JSON(http.StatusOK, gin.H{"article": article})
}
