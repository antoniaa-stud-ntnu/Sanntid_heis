package master



// Connection with elevators
// Needs to update state variable when new information is available - retry if failed
// Sends the information to backup
// Recieves either ACK or NACK from backup that the information is stored
// If NACK, it needs to resend the information
// Run hall_request_assigner
// Send hall_request_assigner output to elevators

// Alive elevators variable, contains ip adresses and port to all alive elevators

// Update all states variable

// Run hall_request_assigner

// Primary init
// Initialize all states
// go sendIPtoPrimaryBroadcast() //Im alive message
// Choose a backup randomly from the alive elevators --> spawn backup.
//

