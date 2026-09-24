package pebble

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"tiyu/models"

	"github.com/cockroachdb/pebble"
)

var (
	writeMux sync.Mutex
)

var ErrDuplicate = errors.New("ticket already exists")

// Record 保存票根识别结果及其防重信息。
type Record struct {
	TicketCategory     string    `json:"ticket_category"`
	TicketMainCategory string    `json:"ticket_main_category"`
	TicketSN           string    `json:"ticket_sn"`
	Seat               string    `json:"seat"`
	HolderName         string    `json:"holder_name"`
	EventDate          string    `json:"event_date"`
	ImageHash          string    `json:"image_hash"`
	UserImageURL       string    `json:"user_image_url"`
	OCRJSON            []byte    `json:"ocr_json"`
	CreatedAt          time.Time `json:"created_at"`
}

// Init 获取统一的 Pebble 数据库连接。
func Init() error {
	_, err := models.GetPebbleDB()
	return err
}

// SaveIfAbsent 在当前进程内原子地检查并保存票根，返回重复原因。
func SaveIfAbsent(record Record) (string, error) {
	if err := Init(); err != nil {
		return "", fmt.Errorf("open ticket pebble: %w", err)
	}
	db, err := models.GetPebbleDB()
	if err != nil {
		return "", fmt.Errorf("get ticket pebble: %w", err)
	}

	writeMux.Lock()
	defer writeMux.Unlock()

	keys := make([]struct {
		kind string
		key  []byte
	}, 0, 2)
	if record.TicketSN != "" {
		keys = append(keys, struct {
			kind string
			key  []byte
		}{kind: "ticket_sn", key: uniqueKey("ticket_sn", strings.TrimSpace(record.TicketSN))})
	}
	if record.Seat != "" && record.EventDate != "" {
		keys = append(keys, struct {
			kind string
			key  []byte
		}{kind: "seat_date", key: uniqueKey("seat_date", record.TicketCategory+"\x00"+record.Seat+"\x00"+record.EventDate)})
	}
	if record.ImageHash != "" {
		keys = append(keys, struct {
			kind string
			key  []byte
		}{kind: "image", key: uniqueKey("image", record.ImageHash)})
	}
	if len(keys) == 0 {
		return "", errors.New("ticket has no usable unique key")
	}

	for _, item := range keys {
		if item.kind == "image" {
			continue
		}
		value, closer, err := db.Get(item.key)
		if err == nil {
			var existing Record
			isRecent := json.Unmarshal(value, &existing) == nil && time.Since(existing.CreatedAt) >= 0 && time.Since(existing.CreatedAt) <= 30*24*time.Hour
			_ = closer.Close()
			if isRecent {
				return item.kind, ErrDuplicate
			}
			continue
		}
		if !errors.Is(err, pebble.ErrNotFound) {
			return "", fmt.Errorf("check ticket pebble key: %w", err)
		}
	}

	if record.ImageHash != "" {
		duplicate, err := hasSimilarTicket(db, record)
		if err != nil {
			return "", err
		}
		if duplicate {
			return "image_info", ErrDuplicate
		}
	}

	value, err := encodeRecord(record)
	if err != nil {
		return "", err
	}
	batch := db.NewBatch()
	defer batch.Close()
	for _, item := range keys {
		if err := batch.Set(item.key, value, pebble.Sync); err != nil {
			return "", fmt.Errorf("set ticket pebble key: %w", err)
		}
	}
	if err := batch.Commit(pebble.Sync); err != nil {
		return "", fmt.Errorf("commit ticket pebble record: %w", err)
	}
	return "", nil
}

// ClearAll 删除 Pebble 中的全部票根防重数据。
func ClearAll() error {
	if err := Init(); err != nil {
		return fmt.Errorf("open ticket pebble: %w", err)
	}
	db, err := models.GetPebbleDB()
	if err != nil {
		return fmt.Errorf("get ticket pebble: %w", err)
	}

	writeMux.Lock()
	defer writeMux.Unlock()

	iter, err := db.NewIter(nil)
	if err != nil {
		return fmt.Errorf("create pebble iterator: %w", err)
	}
	defer iter.Close()

	batch := db.NewBatch()
	defer batch.Close()
	for iter.First(); iter.Valid(); iter.Next() {
		key := append([]byte(nil), iter.Key()...)
		if err := batch.Delete(key, nil); err != nil {
			return fmt.Errorf("delete pebble key: %w", err)
		}
	}
	if err := iter.Error(); err != nil {
		return fmt.Errorf("iterate pebble keys: %w", err)
	}
	if err := batch.Commit(pebble.Sync); err != nil {
		return fmt.Errorf("commit pebble clear: %w", err)
	}
	return nil
}

const imageHashDistanceThreshold = 8

func hasSimilarTicket(db *pebble.DB, record Record) (bool, error) {
	hash, err := strconv.ParseUint(strings.TrimSpace(record.ImageHash), 16, 64)
	if err != nil {
		return false, nil
	}

	iter, err := db.NewIter(nil)
	if err != nil {
		return false, fmt.Errorf("create ticket image iterator: %w", err)
	}
	defer iter.Close()
	for iter.First(); iter.Valid(); iter.Next() {
		if !strings.HasPrefix(string(iter.Key()), "ticket/image/") {
			continue
		}
		var existing Record
		if err := json.Unmarshal(iter.Value(), &existing); err != nil {
			continue
		}
		age := time.Since(existing.CreatedAt)
		if age < 0 || age > 30*24*time.Hour {
			continue
		}
		existingHash, err := strconv.ParseUint(strings.TrimSpace(existing.ImageHash), 16, 64)
		if err != nil || hammingDistance(hash, existingHash) > imageHashDistanceThreshold {
			continue
		}
		if sameTicketFields(record, existing) {
			return true, nil
		}
	}
	if err := iter.Error(); err != nil {
		return false, fmt.Errorf("iterate ticket images: %w", err)
	}
	return false, nil
}

func sameTicketFields(left, right Record) bool {
	return normalize(left.TicketSN) == normalize(right.TicketSN) &&
		normalize(left.Seat) == normalize(right.Seat) &&
		normalize(left.HolderName) == normalize(right.HolderName) &&
		normalize(left.EventDate) == normalize(right.EventDate)
}

func normalize(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func hammingDistance(left, right uint64) int {
	value := left ^ right
	distance := 0
	for value != 0 {
		value &= value - 1
		distance++
	}
	return distance
}

func uniqueKey(kind, value string) []byte {
	digest := sha256.Sum256([]byte(value))
	return []byte("ticket/" + kind + "/" + hex.EncodeToString(digest[:]))
}

func encodeRecord(record Record) ([]byte, error) {
	value, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("marshal ticket pebble record: %w", err)
	}
	return value, nil
}
