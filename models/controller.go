package models

// Controllers : List of controllers to monitor
type Controllers struct {
	Controllers []Controller `json:"controllers"`
}

// Controller : struct used for the Controller connection
type Controller struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Account  string `json:"account"`
	Protocol string `json:"protocol"`
}
