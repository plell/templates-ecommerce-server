package core

func MigrateUp() {

	DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
	// static type tables
	// &Link{},
	)

	insertStaticRecords()
}

func insertStaticRecords() {
	// DB.Exec(`TRUNCATE TABLE user_types`)
	// ut := UserType{Name: "Basic", ID: USER_TYPE_BASIC}
	// DB.Create(&ut)
	// ut = UserType{Name: "Super", ID: USER_TYPE_SUPER}
	// DB.Create(&ut)
	// ut = UserType{Name: "Customer", ID: USER_TYPE_CUSTOMER}
	// DB.Create(&ut)
}

// type Migrator interface {
// 	// AutoMigrate
// 	AutoMigrate(dst ...interface{}) error

// 	// Database
// 	CurrentDatabase() string
// 	FullDataTypeOf(*schema.Field) clause.Expr

// 	// Tables
// 	CreateTable(dst ...interface{}) error
// 	DropTable(dst ...interface{}) error
// 	HasTable(dst interface{}) bool
// 	RenameTable(oldName, newName interface{}) error

// 	// Columns
// 	AddColumn(dst interface{}, field string) error
// 	DropColumn(dst interface{}, field string) error
// 	AlterColumn(dst interface{}, field string) error
// 	HasColumn(dst interface{}, field string) bool
// 	RenameColumn(dst interface{}, oldName, field string) error
// 	MigrateColumn(dst interface{}, field *schema.Field, columnType *sql.ColumnType) error
// 	ColumnTypes(dst interface{}) ([]*sql.ColumnType, error)

// 	// Constraints
// 	CreateConstraint(dst interface{}, name string) error
// 	DropConstraint(dst interface{}, name string) error
// 	HasConstraint(dst interface{}, name string) bool

// 	// Indexes
// 	CreateIndex(dst interface{}, name string) error
// 	DropIndex(dst interface{}, name string) error
// 	HasIndex(dst interface{}, name string) bool
// 	RenameIndex(dst interface{}, oldName, newName string) error
//   }
