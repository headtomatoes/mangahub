-- quickly create an admin user
INSERT INTO users (username, email, password_hash, role) VALUES
    -- user name: admin, pass: SecurePass123!
    ('admin', 'admin@mangahub.com', '$2a$10$r30qt6kx2cRzMxMfu5uVGOUyOrjHujk7MVU8YN.kodndtJmPQMRbW', 'admin');
