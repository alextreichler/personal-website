package repository

func (d *Database) GetSetting(key string) (string, error) {
	var value string
	err := d.Conn.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (d *Database) UpdateSetting(key, value string) error {
	_, err := d.Conn.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", key, value)
	return err
}
