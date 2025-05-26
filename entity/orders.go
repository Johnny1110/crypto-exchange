package entity

type OrderStatus string

const (
	// ORDER_NEW indicates an order that has just been created and not yet matched.
	ORDER_NEW OrderStatus = "NEW"

	// ORDER_PARTIAL indicates an order that has been partially filled.
	ORDER_PARTIAL OrderStatus = "PARTIAL"

	// ORDER_FILLED indicates an order that has been completely filled.
	ORDER_FILLED OrderStatus = "FILLED"

	// ORDER_CANCELED indicates an order that has been canceled.
	ORDER_CANCELED OrderStatus = "CANCELED"
)
