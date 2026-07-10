package controllers

import (
	"bytes"
	"catalog-service/models"
	"catalog-service/repository"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// newUploadCtx fabrique un contexte gin authentifie avec une requete
// multipart/form-data portant un champ "image" (name/content donnes).
func newUploadCtx(t *testing.T, userID uint, articleID uint, fieldName, filename string, content []byte) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if fieldName != "" {
		part, err := writer.CreateFormFile(fieldName, filename)
		if err != nil {
			t.Fatalf("creation champ multipart : %v", err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatalf("ecriture contenu multipart : %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("fermeture writer multipart : %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", float64(userID))
	c.Request = httptest.NewRequest(http.MethodPost, "/", &body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	c.Params = gin.Params{{Key: "id", Value: itoa(articleID)}}
	return c, w
}

var pngSignature = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0}

func TestUploadArticleImageSuccess(t *testing.T) {
	setupCatalogDB(t)
	t.Setenv("UPLOAD_DIR", t.TempDir())
	art := seedArticleForOffer(t, 1, false)

	c, w := newUploadCtx(t, 1, art.ID, "image", "cover.png", pngSignature)
	UploadArticleImage(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status attendu 200, obtenu %d (%s)", w.Code, w.Body.String())
	}
	var reloaded models.Article
	repository.DB.First(&reloaded, art.ID)
	if reloaded.ImageURL == "" {
		t.Error("ImageURL attendu non vide (premiere photo = couverture)")
	}
	if len(reloaded.Images) != 1 {
		t.Errorf("Images attendu avec 1 entree, obtenu %v", reloaded.Images)
	}
}

func TestUploadArticleImageNotFound(t *testing.T) {
	setupCatalogDB(t)
	t.Setenv("UPLOAD_DIR", t.TempDir())

	c, w := newUploadCtx(t, 1, 999, "image", "cover.png", pngSignature)
	UploadArticleImage(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("article inexistant : status attendu 404, obtenu %d", w.Code)
	}
}

func TestUploadArticleImageForbidden(t *testing.T) {
	setupCatalogDB(t)
	t.Setenv("UPLOAD_DIR", t.TempDir())
	art := seedArticleForOffer(t, 7, false)

	c, w := newUploadCtx(t, 1, art.ID, "image", "cover.png", pngSignature)
	UploadArticleImage(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("vendeur different : status attendu 403, obtenu %d", w.Code)
	}
}

func TestUploadArticleImageTooManyImages(t *testing.T) {
	setupCatalogDB(t)
	t.Setenv("UPLOAD_DIR", t.TempDir())
	art := seedArticleForOffer(t, 1, false)

	images := make(models.StringSlice, maxArticleImages)
	for i := range images {
		images[i] = "/uploads/existing.png"
	}
	repository.DB.Model(&art).Update("images", images)

	c, w := newUploadCtx(t, 1, art.ID, "image", "cover.png", pngSignature)
	UploadArticleImage(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("galerie pleine : status attendu 400, obtenu %d", w.Code)
	}
}

func TestUploadArticleImageMissingFile(t *testing.T) {
	setupCatalogDB(t)
	t.Setenv("UPLOAD_DIR", t.TempDir())
	art := seedArticleForOffer(t, 1, false)

	c, w := newUploadCtx(t, 1, art.ID, "", "", nil)
	UploadArticleImage(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("fichier manquant : status attendu 400, obtenu %d", w.Code)
	}
}

func TestUploadArticleImageUnsupportedType(t *testing.T) {
	setupCatalogDB(t)
	t.Setenv("UPLOAD_DIR", t.TempDir())
	art := seedArticleForOffer(t, 1, false)

	c, w := newUploadCtx(t, 1, art.ID, "image", "notes.txt", []byte("ceci n'est pas une image"))
	UploadArticleImage(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("type non supporte : status attendu 400, obtenu %d", w.Code)
	}
}

func TestUploadArticleImageSecondPhotoDoesNotOverwriteCover(t *testing.T) {
	setupCatalogDB(t)
	t.Setenv("UPLOAD_DIR", t.TempDir())
	art := seedArticleForOffer(t, 1, false)

	c, w := newUploadCtx(t, 1, art.ID, "image", "cover.png", pngSignature)
	UploadArticleImage(c)
	if w.Code != http.StatusOK {
		t.Fatalf("premiere photo : status attendu 200, obtenu %d", w.Code)
	}
	var afterFirst models.Article
	repository.DB.First(&afterFirst, art.ID)
	firstCover := afterFirst.ImageURL

	c, w = newUploadCtx(t, 1, art.ID, "image", "gallery.png", pngSignature)
	UploadArticleImage(c)
	if w.Code != http.StatusOK {
		t.Fatalf("deuxieme photo : status attendu 200, obtenu %d", w.Code)
	}

	var reloaded models.Article
	repository.DB.First(&reloaded, art.ID)
	if reloaded.ImageURL != firstCover {
		t.Errorf("la couverture ne doit pas changer, attendu %q obtenu %q", firstCover, reloaded.ImageURL)
	}
	if len(reloaded.Images) != 2 {
		t.Errorf("Images attendu avec 2 entrees, obtenu %v", reloaded.Images)
	}
}
