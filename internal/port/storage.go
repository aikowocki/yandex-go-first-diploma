package port

type Storage interface {
	DB() DB
	TxManager() TxManager
	UserRepo() UserRepository
	OrderRepo() OrderRepository
	BalanceRepo() BalanceRepository
}
