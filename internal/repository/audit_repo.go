package repository

func (d *Database) CreateAuditLog(action, details string) error {
	_, err := d.Conn.Exec("INSERT INTO audit_logs (action, details) VALUES (?, ?)", action, details)
	return err
}
