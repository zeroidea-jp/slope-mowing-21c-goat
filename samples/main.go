package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

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

func send_command_to_each_motors(ctx context.Context, pTx *socketcan.Transmitter, motor_ids []MotorID, s_data []can.Data) error {
	for i, motor_id := range motor_ids {

		err := send_command(ctx, pTx, motor_id, s_data[i], FLAG_SINGLE_MOTOR)
		if err != nil {
			log.Println("error when sending command")
			log.Println(err)
			return err
		}
	}

	return nil
}

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
		frame := pRecv.Frame()
		fmt.Println(frame.String())
		select {
		case <-ctx.Done():
			log.Println("ctx.Done() proceeded.")
			break LOOP
		default:
		}
	}
	pWg.Done()
}

type current_in_deci_A_int16 int16
type angle_in_deci_degree_int16 int16
type angle_in_deci_degree_int32 int32

type shaft_angle_speed_in_deg_per_sec_int16 int16

func (deci_A current_in_deci_A_int16) to_can_data_alley() [2]uint8 {
	return (int16_to_uint8_alley(int16(deci_A)))
	// return integer_to_alley_uint8_x_2(deci_A)
}

func (deci_deg angle_in_deci_degree_int16) to_can_data_alley() [2]uint8 {
	return (int16_to_uint8_alley(int16(deci_deg)))
	// return integer_to_alley_uint8_x_2(deci_deg)
}

func (deci_deg angle_in_deci_degree_int32) to_can_data_alley() [4]uint8 {
	return (int32_to_uint8_alley(int32(deci_deg)))
	// return integer_to_alley_uint8_x_4(deci_deg)
}

func (deg_per_sec shaft_angle_speed_in_deg_per_sec_int16) to_can_data_alley() [2]uint8 {
	return (int16_to_uint8_alley(int16(deg_per_sec)))
	// return integer_to_alley_uint8_x_2(deg_per_sec)
}

func int16_to_uint8_alley(int16_num int16) [2]uint8 {
	return [2]uint8{
		uint8(int16_num & 0xFF),
		uint8(int16_num >> 8 & 0xFF),
	}
}

func int32_to_uint8_alley(int32_num int32) [4]uint8 {
	return [4]uint8{
		uint8(int32_num & 0xFF),
		uint8(int32_num >> 8 & 0xFF),
		uint8(int32_num >> 16 & 0xFF),
		uint8(int32_num >> 24 & 0xFF),
	}
}

// func integer_to_uint8_alley(must_be_integer interface{}) [4]uint8 {
// 	uint8_alley_for_can_data := [4]uint8{0, 0, 0, 0}
// 	if integer, ok := must_be_integer.(int32); ok {
// 		for i, _ := range uint8_alley_for_can_data {
// 			uint8_alley_for_can_data[i] = uint8(integer >> (8 * i) & 0xFF)
// 		}
// 	} else if integer, ok := must_be_integer.(int16); ok {
// 		uint8_alley_for_can_data[0] = uint8(integer & 0xFF)
// 		uint8_alley_for_can_data[0] = uint8(integer >> 8 & 0xFF)
// 	}
// 	return uint8_alley_for_can_data
// }

func main() {

}
