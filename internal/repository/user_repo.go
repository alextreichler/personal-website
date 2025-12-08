package repository

func (d *Database) CountUsers() (int, error) {
	var count int
	err := d.Conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}
