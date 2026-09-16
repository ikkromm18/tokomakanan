CREATE TABLE IF NOT EXISTS package_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    package_id BIGINT UNSIGNED NOT NULL,
    product_id BIGINT UNSIGNED NOT NULL,
    quantity INT UNSIGNED NOT NULL,
    UNIQUE KEY uk_package_items_package_product (package_id, product_id),
    CONSTRAINT fk_package_items_package_id FOREIGN KEY (package_id) REFERENCES product_packages(id) ON DELETE CASCADE,
    CONSTRAINT fk_package_items_product_id FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
