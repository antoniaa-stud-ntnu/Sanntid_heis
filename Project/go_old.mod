module Project-go

go 1.20

require Driver-go v0.0.0
replace Driver-go => ./SingleElevator/Driver-go

import{
	'Project-go/Driver-go/elevio'
}