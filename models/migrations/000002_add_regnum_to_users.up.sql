ALTER TABLE users ADD COLUMN reg_num VARCHAR(10) NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_reg_num_unique UNIQUE (reg_num);