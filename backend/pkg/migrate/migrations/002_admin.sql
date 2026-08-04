-- Chiro — SaaS: roles y estado de cuenta.
-- El admin no es una tabla aparte: es un users con role='admin'.
ALTER TABLE users ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user','admin'));
ALTER TABLE users ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','disabled'));
