export interface IArticle{
    name:string,
    description:string,
    prix:number,
    fraisPort:number,
    categoryId:number,
    photo:string,
    category:ICategorie,
}
export interface ICategorie{
    name:string,
    description:string
}