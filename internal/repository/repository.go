package repository

import (
	"database/sql"
	"errors"
	"time"
	"chat-system-go/internal/database"
	"chat-system-go/internal/models"
)

// ----- Session -----
func CreateSession(sid, token string, expireDays int) error {
	now := time.Now().UnixMilli()
	expire := now + int64(expireDays*24*3600*1000)
	_, err := database.DB.Exec(
		`INSERT INTO sessions (sid, token, created_at, last_active, expire_at) VALUES (?, ?, ?, ?, ?)`,
		sid, token, now, now, expire,
	)
	return err
}

func GetSession(sid string) (*models.Session, error) {
	row := database.DB.QueryRow(`SELECT sid, token, created_at, last_active, expire_at FROM sessions WHERE sid = ?`, sid)
	var s models.Session
	err := row.Scan(&s.Sid, &s.Token, &s.CreatedAt, &s.LastActive, &s.ExpireAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func UpdateSessionLastActive(sid string) error {
	_, err := database.DB.Exec(`UPDATE sessions SET last_active = ? WHERE sid = ?`, time.Now().UnixMilli(), sid)
	return err
}

func DeleteSession(sid string) error {
	_, err := database.DB.Exec(`DELETE FROM sessions WHERE sid = ?`, sid)
	return err
}

func ListExpiredSessions(expireBefore int64) ([]string, error) {
	rows, err := database.DB.Query(`SELECT sid FROM sessions WHERE expire_at < ?`, expireBefore)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sids []string
	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err != nil {
			return nil, err
		}
		sids = append(sids, sid)
	}
	return sids, nil
}

// ----- Message -----
func InsertMessage(msg *models.Message) error {
	_, err := database.DB.Exec(
		`INSERT INTO messages (id, sid, role, type, text, file_name, file_size, file_url, mime_type, time, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.Sid, msg.Role, msg.Type, msg.Text, msg.FileName, msg.FileSize, msg.FileURL, msg.MimeType, msg.Time, msg.Metadata,
	)
	return err
}

func GetMessagesBySidAfter(sid string, after int64, limit int) ([]models.Message, error) {
	rows, err := database.DB.Query(
		`SELECT id, role, type, text, file_name, file_size, file_url, mime_type, time, metadata
		FROM messages WHERE sid = ? AND time > ? ORDER BY time ASC LIMIT ?`,
		sid, after, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessages(rows)
}

func GetMessagesBySidBefore(sid string, before int64, limit int) ([]models.Message, error) {
	rows, err := database.DB.Query(
		`SELECT id, role, type, text, file_name, file_size, file_url, mime_type, time, metadata
		FROM messages WHERE sid = ? AND time < ? ORDER BY time DESC LIMIT ?`,
		sid, before, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	msgs, err := scanMessages(rows)
	if err != nil {
		return nil, err
	}
	// reverse to ascending
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

func scanMessages(rows *sql.Rows) ([]models.Message, error) {
	var msgs []models.Message
	for rows.Next() {
		var m models.Message
		err := rows.Scan(&m.ID, &m.Role, &m.Type, &m.Text, &m.FileName, &m.FileSize, &m.FileURL, &m.MimeType, &m.Time, &m.Metadata)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func CountMessagesBySidAndRole(sid, role string) (int, error) {
	var count int
	err := database.DB.QueryRow(`SELECT COUNT(*) FROM messages WHERE sid = ? AND role = ?`, sid, role).Scan(&count)
	return count, err
}

func GetLastUserMessageTime(sid string) (int64, error) {
	var t int64
	err := database.DB.QueryRow(`SELECT time FROM messages WHERE sid = ? AND role = 'user' ORDER BY time ASC LIMIT 1`, sid).Scan(&t)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return t, err
}

// ----- File -----
func InsertFileRecord(sid, path string, size int64) error {
	_, err := database.DB.Exec(`INSERT INTO files (sid, path, size, created_at) VALUES (?, ?, ?, ?)`, sid, path, size, time.Now().UnixMilli())
	return err
}

func GetFilesBySid(sid string) ([]models.FileRecord, error) {
	rows, err := database.DB.Query(`SELECT id, path, size, created_at FROM files WHERE sid = ?`, sid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var files []models.FileRecord
	for rows.Next() {
		var f models.FileRecord
		if err := rows.Scan(&f.ID, &f.Path, &f.Size, &f.CreatedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

func DeleteFilesBySid(sid string) error {
	_, err := database.DB.Exec(`DELETE FROM files WHERE sid = ?`, sid)
	return err
}

// ----- Topic -----
func GetTopic(sid string) (*models.Topic, error) {
	row := database.DB.QueryRow(`SELECT sid, topic_id, topic_name FROM topics WHERE sid = ?`, sid)
	var t models.Topic
	err := row.Scan(&t.Sid, &t.TopicID, &t.TopicName)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func InsertTopic(topic *models.Topic) error {
	_, err := database.DB.Exec(`INSERT INTO topics (sid, topic_id, topic_name) VALUES (?, ?, ?)`, topic.Sid, topic.TopicID, topic.TopicName)
	return err
}

func DeleteTopic(sid string) error {
	_, err := database.DB.Exec(`DELETE FROM topics WHERE sid = ?`, sid)
	return err
}

func GetSidByTopicID(topicID int64) (string, error) {
	var sid string
	err := database.DB.QueryRow(`SELECT sid FROM topics WHERE topic_id = ?`, topicID).Scan(&sid)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return sid, err
}

// ----- DeviceName -----
func GetDeviceName(sid string) (*models.DeviceName, error) {
	row := database.DB.QueryRow(`SELECT sid, device_type, seq FROM device_names WHERE sid = ?`, sid)
	var d models.DeviceName
	err := row.Scan(&d.Sid, &d.DeviceType, &d.Seq)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &d, err
}

func InsertDeviceName(dev *models.DeviceName) error {
	_, err := database.DB.Exec(`INSERT INTO device_names (sid, device_type, seq) VALUES (?, ?, ?)`, dev.Sid, dev.DeviceType, dev.Seq)
	return err
}

func DeleteDeviceName(sid string) error {
	_, err := database.DB.Exec(`DELETE FROM device_names WHERE sid = ?`, sid)
	return err
}

// ----- RateLimit -----
func GetRateLimit(ip string, windowStart int64) (*models.RateLimit, error) {
	row := database.DB.QueryRow(`SELECT ip, window_start, count, updated_at FROM rate_limits WHERE ip = ? AND window_start = ?`, ip, windowStart)
	var rl models.RateLimit
	err := row.Scan(&rl.IP, &rl.WindowStart, &rl.Count, &rl.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &rl, err
}

func UpsertRateLimit(rl *models.RateLimit) error {
	_, err := database.DB.Exec(
		`INSERT INTO rate_limits (ip, window_start, count, updated_at) VALUES (?, ?, ?, ?)
		ON CONFLICT(ip, window_start) DO UPDATE SET count = excluded.count, updated_at = excluded.updated_at`,
		rl.IP, rl.WindowStart, rl.Count, rl.UpdatedAt,
	)
	return err
}

// ----- BlockedIP -----
func GetBlockedIP(ip string) (*models.BlockedIP, error) {
	row := database.DB.QueryRow(`SELECT ip, blocked_at, expire_at FROM blocked_ips WHERE ip = ?`, ip)
	var b models.BlockedIP
	err := row.Scan(&b.IP, &b.BlockedAt, &b.ExpireAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &b, err
}

func UpsertBlockedIP(ip string, blockedAt, expireAt int64) error {
	_, err := database.DB.Exec(
		`INSERT INTO blocked_ips (ip, blocked_at, expire_at) VALUES (?, ?, ?)
		ON CONFLICT(ip) DO UPDATE SET blocked_at = excluded.blocked_at, expire_at = excluded.expire_at`,
		ip, blockedAt, expireAt,
	)
	return err
}

func DeleteExpiredBlockedIPs(now int64) error {
	_, err := database.DB.Exec(`DELETE FROM blocked_ips WHERE expire_at < ?`, now)
	return err
}

// ----- AutoReplyFlag -----
func GetAutoReplyFlag(sid string) (*models.AutoReplyFlag, error) {
	row := database.DB.QueryRow(`SELECT sid, sent_at FROM autoreply_flags WHERE sid = ?`, sid)
	var f models.AutoReplyFlag
	err := row.Scan(&f.Sid, &f.SentAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &f, err
}

func InsertAutoReplyFlag(sid string, sentAt int64) error {
	_, err := database.DB.Exec(`INSERT INTO autoreply_flags (sid, sent_at) VALUES (?, ?)`, sid, sentAt)
	return err
}

func DeleteAutoReplyFlag(sid string) error {
	_, err := database.DB.Exec(`DELETE FROM autoreply_flags WHERE sid = ?`, sid)
	return err
}

// ----- GlobalSeq -----
func GetGlobalSeq(key string) (int64, error) {
	var val int64
	err := database.DB.QueryRow(`SELECT value FROM global_seq WHERE key = ?`, key).Scan(&val)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return val, err
}

func IncrementGlobalSeq(key string) (int64, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var val int64
	err = tx.QueryRow(`SELECT value FROM global_seq WHERE key = ?`, key).Scan(&val)
	if err == sql.ErrNoRows {
		val = 0
		_, err = tx.Exec(`INSERT INTO global_seq (key, value) VALUES (?, ?)`, key, 1)
		if err != nil {
			return 0, err
		}
		val = 1
	} else if err != nil {
		return 0, err
	} else {
		val++
		_, err = tx.Exec(`UPDATE global_seq SET value = ? WHERE key = ?`, val, key)
		if err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return val, nil
}
