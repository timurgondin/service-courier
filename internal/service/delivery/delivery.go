package delivery

type Service struct {
	deliveryRepo     deliveryRepository
	courierRepo      courierRepository
	transportFactory TransportFactory
	txManager        transactionManager
	clock            Clock
}

func NewDeliveryService(
	deliveryRepo deliveryRepository,
	courierRepo courierRepository,
	transportFactory TransportFactory,
	txManager transactionManager,
	clock Clock,
) *Service {
	return &Service{
		deliveryRepo:     deliveryRepo,
		courierRepo:      courierRepo,
		transportFactory: transportFactory,
		txManager:        txManager,
		clock:            clock,
	}

}
