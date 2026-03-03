package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"productsManager/internal/products/models"
)

type ProductRepo struct {
	conn *pgxpool.Pool
}

func NewProductRepo(conn *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{conn: conn}
}

func (r *ProductRepo) CreateProduct(ctx context.Context, name string, price int) (models.Product, error) {
	const query = `
		INSERT INTO products (name, price)
		VALUES ($1, $2)
		RETURNING id, name, price, created_at
	`

	var product models.Product
	if err := r.conn.QueryRow(ctx, query, name, price).Scan(
		&product.ID,
		&product.Name,
		&product.Price,
		&product.CreatedAt,
	); err != nil {
		return models.Product{}, fmt.Errorf("insert product: %w", err)
	}

	return product, nil
}

func (r *ProductRepo) DeleteProduct(ctx context.Context, id int64) (bool, error) {
	tag, err := r.conn.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete product: %w", err)
	}

	return tag.RowsAffected() > 0, nil
}

func (r *ProductRepo) ListProducts(ctx context.Context, page int, limit int) ([]models.Product, int64, error) {
	var total int64
	if err := r.conn.QueryRow(ctx, `SELECT count(*) FROM products`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	rows, err := r.conn.Query(ctx, `
		SELECT id, name, price, created_at
		FROM products
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	items := make([]models.Product, 0, limit)
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan product: %w", err)
		}
		items = append(items, product)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate products: %w", err)
	}

	return items, total, nil
}

func (r *ProductRepo) TruncateProducts(ctx context.Context) error {
	if _, err := r.conn.Exec(ctx, `TRUNCATE TABLE products RESTART IDENTITY`); err != nil {
		return fmt.Errorf("truncate products: %w", err)
	}

	return nil
}
