package dao

import "gorm.io/gorm"

const (
	bookTableName = "book_upload_maintenance_field"
	bookCardSort = 6
)

type BookResult struct {
	BookSign string
	ApiKey string
}

// SearchBook 查询所有非定制的绘本数据
func SearchBook() []BookResult{
	results := make([]BookResult, 0)
	DB.Table(bookTableName).Select("book_sign,api_key").Where("allow_custom = ? and book_sort <> ? ", 0, bookCardSort).Scan(results)
	return results
}


// SearchCard 查询所有非定制的绘本数据
func SearchCard() []BookResult{
	results := make([]BookResult, 0)
	DB.Table(bookTableName).Select("book_sign,api_key").Where("allow_custom = ? and book_sort = ? ", 0, bookCardSort).Scan(results)
	return results
}


// SearchUnCustomBook 查询非定制的绘本数据
func SearchUnCustomBook() []BookResult{
	results := make([]BookResult, 0)
	tx := DB.Session(&gorm.Session{}).Table(bookTableName).Select("book_sign,api_key")
	tx.Where("allow_custom = ? and book_sort = ?  and attribute = ?", 0, bookCardSort, 0).Scan(results)
	return results
}




