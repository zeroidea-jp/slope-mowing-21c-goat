package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.einride.tech/can"
	"go.einride.tech/can/pkg/socketcan"
)

func main() {
	var wg sync.WaitGroup

	ctx := context.TODO()

	pConn, err := socketcan.DialContext(ctx, "can", "can0")
	if err != nil {
		log.Println("Error when Dialialing with Context.")
		log.Fatal(err)
	}
	defer pConn.Close()

	pRecv := socketcan.NewReceiver(pConn)
	defer pRecv.Close()

	wg.Add(1)
	go func(ctx context.Context, recv *socketcan.Receiver) {
	LOOP:
		for recv.Receive() {
			frame := recv.Frame()
			fmt.Println(frame.String())
			select {
			case <-ctx.Done():
				log.Println("ctx.Done() proceeded.")
				break LOOP
			default:
			}
		}
	}(ctx, pRecv)
	wg.Done()

	data_to_torque_closed_loop_control := can.Data{0xA1, 0x00, 0x00, 0x00, 0x0b, 0x00, 0x00, 0x00}
	data_to_shutdown_motor := can.Data{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	// data := can.Data{0x30, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	motor_id := uint32(0x141)

	frame_to_torque_closed_loop_control := can.Frame{
		ID:         motor_id,
		Length:     8,
		Data:       data_to_torque_closed_loop_control,
		IsRemote:   false,
		IsExtended: false,
	}

	frame_to_shutdown_motor := can.Frame{
		ID:         motor_id,
		Length:     8,
		Data:       data_to_shutdown_motor,
		IsRemote:   false,
		IsExtended: false,
	}

	pTx := socketcan.NewTransmitter(pConn)
	err = pTx.TransmitFrame(ctx, frame_to_torque_closed_loop_control)
	if err != nil {
		log.Println("Error when transmitting Frame (1)")
	}

	time.Sleep(time.Second)

	err = pTx.TransmitFrame(ctx, frame_to_shutdown_motor)
	if err != nil {
		log.Println("Error when transmitting Frame (2)")
	}

	defer pTx.Close()

	wg.Wait()

	// fmt.Printf("Get message: %v\n", msg)
}
