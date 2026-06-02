package repository

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/alexkalak/whatever-system/src/modules/trades/dexactions/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const defaultDexActionOrder = "block_number DESC, index_in_block DESC, index_in_tx DESC, id ASC"

type cursor struct {
	SortValue    string
	BlockNumber  uint64
	IndexInBlock uint64
	IndexInTx    uint64
	ID           uuid.UUID
}

func selectWithSeenInMempool(table string) string {
	return table + ".*, EXISTS (SELECT 1 FROM mempool_hashes mh WHERE mh.hash = " + table + ".tx_hash AND mh.chain_id = " + table + ".chain_id) AS seen_in_mempool"
}

func parseCursor(value string) (*cursor, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parts := strings.Split(value, ":")
	if len(parts) != 5 {
		return nil, fmt.Errorf("invalid cursor")
	}
	blockNumber, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor block number")
	}
	indexInBlock, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor index")
	}
	indexInTx, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor tx index")
	}
	id, err := uuid.Parse(parts[4])
	if err != nil {
		return nil, fmt.Errorf("invalid cursor id")
	}
	return &cursor{SortValue: parts[0], BlockNumber: blockNumber, IndexInBlock: indexInBlock, IndexInTx: indexInTx, ID: id}, nil
}

func formatCursor(sortValue string, blockNumber, indexInBlock, indexInTx uint64, id uuid.UUID) string {
	if sortValue == "" {
		sortValue = "_"
	}
	return fmt.Sprintf("%s:%d:%d:%d:%s", sortValue, blockNumber, indexInBlock, indexInTx, id.String())
}

func orderClause(orderBy, direction string) string {
	if direction != "asc" {
		direction = "desc"
	}
	sqlDirection := "DESC"
	if direction == "asc" {
		sqlDirection = "ASC"
	}
	switch orderBy {
	case "amount0":
		return "ABS(amount0) " + sqlDirection + ", " + defaultDexActionOrder
	case "amount1":
		return "ABS(amount1) " + sqlDirection + ", " + defaultDexActionOrder
	default:
		if direction == "asc" {
			return "block_number ASC, index_in_block ASC, index_in_tx ASC, id ASC"
		}
		return defaultDexActionOrder
	}
}

func timeCursorSQL(direction string) string {
	if direction == "asc" {
		return "block_number > ? OR (block_number = ? AND index_in_block > ?) OR (block_number = ? AND index_in_block = ? AND index_in_tx > ?) OR (block_number = ? AND index_in_block = ? AND index_in_tx = ? AND id > ?)"
	}
	return "block_number < ? OR (block_number = ? AND index_in_block < ?) OR (block_number = ? AND index_in_block = ? AND index_in_tx < ?) OR (block_number = ? AND index_in_block = ? AND index_in_tx = ? AND id > ?)"
}

func timeCursorArgs(c *cursor) []any {
	return []any{c.BlockNumber, c.BlockNumber, c.IndexInBlock, c.BlockNumber, c.IndexInBlock, c.IndexInTx, c.BlockNumber, c.IndexInBlock, c.IndexInTx, c.ID}
}

func applyCursor(query *gorm.DB, c *cursor, orderBy, direction string) *gorm.DB {
	if c == nil {
		return query
	}
	if orderBy == "amount0" || orderBy == "amount1" {
		expr := "ABS(amount0)"
		if orderBy == "amount1" {
			expr = "ABS(amount1)"
		}
		op := "<"
		if direction == "asc" {
			op = ">"
		}
		where := fmt.Sprintf("%s %s ? OR (%s = ? AND (%s))", expr, op, expr, timeCursorSQL("desc"))
		args := append([]any{c.SortValue, c.SortValue}, timeCursorArgs(c)...)
		return query.Where(where, args...)
	}
	return query.Where(timeCursorSQL(direction), timeCursorArgs(c)...)
}

func bigintAbsString(value *big.Int) string {
	if value == nil {
		return "0"
	}
	return new(big.Int).Abs(value).String()
}

type UniswapV2Repository interface {
	Create(action *entities.DexActionUniswapV2) error
	GetPaginated(page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, error)
	GetByID(id uuid.UUID) (*entities.DexActionUniswapV2, error)
	GetByTxHash(txHash string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, error)
	GetByChainIDAndDexAddress(chainID uint64, dexAddress string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, error)
	GetByChainIDAndDexAddressCursor(chainID uint64, dexAddress string, cursor string, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, string, error)
	GetByChainIDPoolTxHashAndIndexes(chainID uint64, poolAddress, txHash, actionType string, indexInBlock, indexInTx uint64) (*entities.DexActionUniswapV2, error)
}

type UniswapV3Repository interface {
	Create(action *entities.DexActionUniswapV3) error
	GetPaginated(page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, error)
	GetByID(id uuid.UUID) (*entities.DexActionUniswapV3, error)
	GetByTxHash(txHash string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, error)
	GetByChainIDAndDexAddress(chainID uint64, dexAddress string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, error)
	GetByChainIDAndDexAddressCursor(chainID uint64, dexAddress string, cursor string, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, string, error)
	GetByChainIDPoolTxHashActionAndIndexes(chainID uint64, poolAddress, txHash, actionType string, indexInBlock, indexInTx uint64) (*entities.DexActionUniswapV3, error)
}

type uniswapV2Repository struct{ db *gorm.DB }
type uniswapV3Repository struct{ db *gorm.DB }

func NewUniswapV2Repository(db *gorm.DB) UniswapV2Repository { return &uniswapV2Repository{db: db} }
func NewUniswapV3Repository(db *gorm.DB) UniswapV3Repository { return &uniswapV3Repository{db: db} }

func (r *uniswapV2Repository) base() *gorm.DB { return r.db.Model(&entities.DexActionUniswapV2{}) }
func (r *uniswapV3Repository) base() *gorm.DB { return r.db.Model(&entities.DexActionUniswapV3{}) }

func (r *uniswapV2Repository) Create(action *entities.DexActionUniswapV2) error {
	return r.db.Create(action).Error
}
func (r *uniswapV3Repository) Create(action *entities.DexActionUniswapV3) error {
	return r.db.Create(action).Error
}

func (r *uniswapV2Repository) GetPaginated(page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, error) {
	var actions []entities.DexActionUniswapV2
	var total int64
	if err := r.base().Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.base().Select(selectWithSeenInMempool("dex_action_uniswap_v2")).Order(orderClause(orderBy, direction)).Offset((page - 1) * limit).Limit(limit).Find(&actions).Error
	return actions, total, err
}
func (r *uniswapV3Repository) GetPaginated(page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, error) {
	var actions []entities.DexActionUniswapV3
	var total int64
	if err := r.base().Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.base().Select(selectWithSeenInMempool("dex_action_uniswap_v3")).Order(orderClause(orderBy, direction)).Offset((page - 1) * limit).Limit(limit).Find(&actions).Error
	return actions, total, err
}
func (r *uniswapV2Repository) GetByID(id uuid.UUID) (*entities.DexActionUniswapV2, error) {
	var a entities.DexActionUniswapV2
	err := r.base().Select(selectWithSeenInMempool("dex_action_uniswap_v2")).First(&a, "id = ?", id).Error
	return &a, err
}
func (r *uniswapV3Repository) GetByID(id uuid.UUID) (*entities.DexActionUniswapV3, error) {
	var a entities.DexActionUniswapV3
	err := r.base().Select(selectWithSeenInMempool("dex_action_uniswap_v3")).First(&a, "id = ?", id).Error
	return &a, err
}

func (r *uniswapV2Repository) GetByTxHash(txHash string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, error) {
	var a []entities.DexActionUniswapV2
	var total int64
	q := r.base().Where("LOWER(tx_hash) = ?", txHash)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Select(selectWithSeenInMempool("dex_action_uniswap_v2")).Order(orderClause(orderBy, direction)).Offset((page - 1) * limit).Limit(limit).Find(&a).Error
	return a, total, err
}
func (r *uniswapV3Repository) GetByTxHash(txHash string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, error) {
	var a []entities.DexActionUniswapV3
	var total int64
	q := r.base().Where("LOWER(tx_hash) = ?", txHash)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Select(selectWithSeenInMempool("dex_action_uniswap_v3")).Order(orderClause(orderBy, direction)).Offset((page - 1) * limit).Limit(limit).Find(&a).Error
	return a, total, err
}

func (r *uniswapV2Repository) GetByChainIDAndDexAddress(chainID uint64, dexAddress string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, error) {
	var a []entities.DexActionUniswapV2
	var total int64
	q := r.base().Where("chain_id = ? AND LOWER(dex_address) = ?", chainID, dexAddress)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.
		Select(selectWithSeenInMempool("dex_action_uniswap_v2")).
		Order(orderClause(orderBy, direction)).
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&a).Error

	return a, total, err
}
func (r *uniswapV3Repository) GetByChainIDAndDexAddress(chainID uint64, dexAddress string, page, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, error) {
	var a []entities.DexActionUniswapV3
	var total int64
	q := r.base().Where("chain_id = ? AND LOWER(dex_address) = ?", chainID, dexAddress)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.
		Select(selectWithSeenInMempool("dex_action_uniswap_v3")).
		Order(orderClause(orderBy, direction)).
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&a).Error

	return a, total, err
}

func (r *uniswapV2Repository) GetByChainIDAndDexAddressCursor(chainID uint64, dexAddress string, cursorValue string, limit int, orderBy, direction string) ([]entities.DexActionUniswapV2, int64, string, error) {
	var a []entities.DexActionUniswapV2
	var total int64
	c, err := parseCursor(cursorValue)
	if err != nil {
		return nil, 0, "", err
	}
	q := r.base().Where("chain_id = ? AND LOWER(dex_address) = ?", chainID, dexAddress)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, "", err
	}
	err = applyCursor(q, c, orderBy, direction).
		Select(selectWithSeenInMempool("dex_action_uniswap_v2")).
		Order(orderClause(orderBy, direction)).
		Limit(limit + 1).
		Find(&a).Error

	if err != nil {
		return nil, 0, "", err
	}
	next := ""
	if len(a) > limit {
		a = a[:limit]
		last := a[len(a)-1]
		sort := ""
		if orderBy == "amount0" {
			sort = bigintAbsString(last.Amount0.Int)
		} else if orderBy == "amount1" {
			sort = bigintAbsString(last.Amount1.Int)
		}
		next = formatCursor(sort, last.BlockNumber, last.IndexInBlock, last.IndexInTx, last.ID)
	}

	return a, total, next, nil
}
func (r *uniswapV3Repository) GetByChainIDAndDexAddressCursor(chainID uint64, dexAddress string, cursorValue string, limit int, orderBy, direction string) ([]entities.DexActionUniswapV3, int64, string, error) {
	var a []entities.DexActionUniswapV3
	var total int64
	c, err := parseCursor(cursorValue)
	if err != nil {
		return nil, 0, "", err
	}
	q := r.base().Where("chain_id = ? AND LOWER(dex_address) = ?", chainID, dexAddress)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, "", err
	}
	err = applyCursor(q, c, orderBy, direction).Select(selectWithSeenInMempool("dex_action_uniswap_v3")).Order(orderClause(orderBy, direction)).Limit(limit + 1).Find(&a).Error
	if err != nil {
		return nil, 0, "", err
	}
	next := ""
	if len(a) > limit {
		a = a[:limit]
		last := a[len(a)-1]
		sort := ""
		if orderBy == "amount0" {
			sort = bigintAbsString(last.Amount0.Int)
		} else if orderBy == "amount1" {
			sort = bigintAbsString(last.Amount1.Int)
		}
		next = formatCursor(sort, last.BlockNumber, last.IndexInBlock, last.IndexInTx, last.ID)
	}
	return a, total, next, nil
}

func (r *uniswapV2Repository) GetByChainIDPoolTxHashAndIndexes(chainID uint64, poolAddress, txHash, actionType string, indexInBlock, indexInTx uint64) (*entities.DexActionUniswapV2, error) {
	var a entities.DexActionUniswapV2
	err := r.db.First(&a, "chain_id = ? AND pool_address = ? AND tx_hash = ? AND action_type = ? AND index_in_block = ? AND index_in_tx = ?", chainID, poolAddress, txHash, actionType, indexInBlock, indexInTx).Error
	return &a, err
}
func (r *uniswapV3Repository) GetByChainIDPoolTxHashActionAndIndexes(chainID uint64, poolAddress, txHash, actionType string, indexInBlock, indexInTx uint64) (*entities.DexActionUniswapV3, error) {
	var a entities.DexActionUniswapV3
	err := r.db.First(&a, "chain_id = ? AND pool_address = ? AND tx_hash = ? AND action_type = ? AND index_in_block = ? AND index_in_tx = ?", chainID, poolAddress, txHash, actionType, indexInBlock, indexInTx).Error
	return &a, err
}
