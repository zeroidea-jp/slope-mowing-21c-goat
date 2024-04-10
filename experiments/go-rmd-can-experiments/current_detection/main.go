package main

import (
	"context"
	"log"

	"go.einride.tech/can"
	"go.einride.tech/can/pkg/socketcan"
)

const MOTOR_ID_1 = MotorID(0x01)
const MOTOR_ID_2 = MotorID(0x02)

const FLAG_SINGLE_MOTOR MotorTargetFlag = 0x140

// const CURRENT_CONTRL_VAL

// const FLAG_MULTI_MOTOR MotorTargetFlag = 0x280

type MotorTargetFlag uint32

type MotorModelParameters struct {
}

type MotorID uint8
type Model string

// type MotorModels map[Model]struct {
// }

type Orders map[MotorID]*OrderData

type OrderData struct {
	canFrame can.Frame
}

func main() {
	// 制御する全てのモーターの ID を Slice に格納
	motor_ids := []MotorID{MOTOR_ID_1, MOTOR_ID_2}

	// Context を作成
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// // シグナルを受信するチャンネルを作成
	// sigCh := make(chan os.Signal, 1)
	// signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// go func() {
	// 	<-sigCh
	// 	cancel()
	// }()

	// CAN 通信のインターフェースを開く
	pConn, err := socketcan.DialContext(ctx, "can", "can0")
	if err != nil {
		log.Println("Error when Dialialing with Context.")
		log.Fatal(err)
	}
	defer pConn.Close()

	// 一定のインターバルで CAN 通信を問い合わせる
	pTx := socketcan.NewTransmitter(pConn)
	defer pTx.Close()

	pRx := socketcan.NewReceiver(pConn)
	defer pRx.Close()

	log.Println("Exiting main()")

	// fmt.Printf("Get message: %v\n", msg)
}
