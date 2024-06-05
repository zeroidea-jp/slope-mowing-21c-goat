package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"go.einride.tech/can"
	"go.einride.tech/can/pkg/socketcan"
)

const FLAG_MOTOR_SETUP MotorTargetFlag = 0x300

type MotorTargetFlag uint32

type MotorModelParameters struct {
}

type MotorID uint8
type Model string

func send_command(ctx context.Context, pTx *socketcan.Transmitter, motor_id MotorID, data can.Data, flag MotorTargetFlag) error {
	data_frame := can.Frame{
		ID:         uint32(motor_id) | uint32(flag),
		Length:     8,
		Data:       data,
		IsRemote:   false,
		IsExtended: false,
	}

	err := pTx.TransmitFrame(ctx, data_frame)
	if err != nil {
		log.Println("Error when transmitting Frame (1)")
	}
	return err

}

func receive_reply(ctx context.Context, pConn net.Conn, pWg *sync.WaitGroup) {
	pRecv := socketcan.NewReceiver(pConn)
	defer pRecv.Close()
LOOP:
	for pRecv.Receive() {
		log.Println("Waiting for a reply...")
		frame := pRecv.Frame()
		fmt.Println(frame.String())
		select {
		case <-ctx.Done():
			log.Println("ctx.Done() proceeded.")
			break LOOP
		default:
		}
	}
	log.Println("Closed CAN message receiver.")
	pWg.Done()
}

func main() {

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	var wg sync.WaitGroup
	defer wg.Wait()

	pConn, err := socketcan.DialContext(ctx, "can", "can0")
	if err != nil {
		log.Println("Error when Dialialing with Context.")
		log.Fatal(err)
	}
	defer pConn.Close()

	pTx := socketcan.NewTransmitter(pConn)
	defer pTx.Close()

	wg.Add(1)
	go receive_reply(ctx, pConn, &wg)

	log.Println("Seinding Data to Read ")
	dummy_ID := MotorID(0)
	data_to_read_motor_id := can.Data{0x79, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00}
	err = send_command(ctx, pTx, dummy_ID, data_to_read_motor_id, FLAG_MOTOR_SETUP)
	if err != nil {
		log.Println("error when sending command")
		log.Fatal(err)
	}

	log.Println("Waiting for seconds to get reply...")
	time.Sleep(2 * time.Second)
	log.Println("Exiting main()")

}
