package user

// MockSetMongoInfo is a mock function for testing
var MockSetMongoInfo func(u *User) error

// MockFindByUuid is a mock function for testing
var MockFindByUuid func(uuid string) (*User, error)
