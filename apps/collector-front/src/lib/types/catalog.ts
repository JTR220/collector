export interface Category {
	ID: number;
	CreatedAt?: string;
	UpdatedAt?: string;
	name: string;
	description: string;
}

export interface Article {
	ID: number;
	CreatedAt?: string;
	UpdatedAt?: string;
	name: string;
	description: string;
	prix: number;
	fraisPort: number;
	categoryId: number;
	category?: Category;
}
