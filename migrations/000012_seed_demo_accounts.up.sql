INSERT INTO users (name, email, password_hash, role, is_active, created_at, updated_at)
VALUES 
    ('Budi Owner', 'budi.owner@tokomakanan.com', '$2a$12$DIR4Cy/dx/MkdMapZp/Bn.JsqJrRFbDIFsjzzfzkgmdW91tvOgrLi', 'owner', 1, NOW(), NOW()),
    ('Sarah Kasir', 'kasir1@tokomakanan.com', '$2a$12$jRNc3JX2koV9ISv2BqcXu.3loVXN0U7UXYmKbv2ZxjJMJJM92f4iW', 'admin', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE 
    name = VALUES(name),
    password_hash = VALUES(password_hash),
    role = VALUES(role),
    is_active = VALUES(is_active),
    updated_at = NOW();
