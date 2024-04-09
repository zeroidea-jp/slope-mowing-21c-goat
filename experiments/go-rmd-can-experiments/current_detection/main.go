package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

func recieve_reply_and_save_status_logs(ctx context.Context, pConn net.Conn, pWg *sync.WaitGroup) {
	pRecv := socketcan.NewReceiver(pConn)
	defer pRecv.Close()

LOOP:
	for pRecv.Receive() {
		frame := pRecv.Frame()
		logFrame(frame)
		select {
		case <-ctx.Done():
			log.Println("ctx.Done() proceeded.")
			break LOOP
		default:
		}
	}
	pWg.Done()
}

func logFrame(frame can.Frame) {
	// フレームのData[0]が0x9A, 0x9C, 0x9Dのいずれかである場合にのみログを取る
	if frame.Data[0] == 0x9A || frame.Data[0] == 0x9C || frame.Data[0] == 0x9D {
		// ログファイル名を生成
		filename := fmt.Sprintf("can_log_%X_%X.log", frame.ID, frame.Data[0])

		// ログファイルを開く（または作成する）
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		// フレームの情報をログファイルに書き込む
		log.SetOutput(file)
		log.Printf("ID: %X, Data: %X\n", frame.ID, frame.Data)
	} else {
		fmt.Printf("ID: %X, Data: %X\n", frame.ID, frame.Data)
	}
}

func ask_status_at_interval(ctx context.Context, pTx *socketcan.Transmitter, motor_ids []MotorID, s_data []can.Data, pWg *sync.WaitGroup) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Shutting down...")
			return
		case <-ticker.C:
			send_command_to_each_motors(ctx, pTx, motor_ids, s_data)
			err := send_command_to_each_motors(ctx, pTx, motor_ids, s_data)
			if err != nil {
				log.Println("error when sending read_status command")
				log.Fatal(err)
			}

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
	// 制御する全てのモーターの ID を Slice に格納
	motor_ids := []MotorID{MOTOR_ID_1, MOTOR_ID_2}

	// Context を作成
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// シグナルを受信するチャンネルを作成
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	// WaitGroup を作成
	var wg sync.WaitGroup
	defer wg.Wait()

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

	// go-routine: データを受信し、ステータスに関する問い合わせのみ保存する。
	wg.Add(1)
	go recieve_reply_and_save_status_logs(ctx, pConn, &wg)

	// go-routine: 一定のインターバルでステータスに対する問い合わせをする。
	data_to_read_motor_status_2 := can.Data{0x9C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	s_read_status_data_2 := []can.Data{
		data_to_read_motor_status_2,
		data_to_read_motor_status_2,
	}
	wg.Add(1)
	go ask_status_at_interval(ctx, pTx, motor_ids, s_read_status_data_2, &wg)

	data_to_read_multi_turn_encoder_position_data := can.Data{0x60, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	s_data_to_read_multi_turn_motor_position := []can.Data{
		data_to_read_multi_turn_encoder_position_data,
		data_to_read_multi_turn_encoder_position_data,
	}

	shaft_angle_speed := shaft_angle_speed_in_deg_per_sec_int16(180)
	a_speed := shaft_angle_speed.to_can_data_alley()

	delta_deci_degree := angle_in_deci_degree_int32(0 * 100)
	a_delt_c_deg := delta_deci_degree.to_can_data_alley()
	log.Println(a_delt_c_deg)

	data_to_control_absolute_position_in_closed_loop_0 := can.Data{0xA4, 0x00, a_speed[0], a_speed[1], a_delt_c_deg[0], a_delt_c_deg[1], a_delt_c_deg[2], a_delt_c_deg[3]} // works

	// s_data_to_control_absolute_position_in_closed_loop_0 := []can.Data{
	// 	data_to_control_absolute_position_in_closed_loop_0,
	// 	data_to_control_absolute_position_in_closed_loop_0,
	// }

	delta_deci_degree = angle_in_deci_degree_int32(-360 * 100)
	a_delt_c_deg = delta_deci_degree.to_can_data_alley()
	log.Println(a_delt_c_deg)
	data_to_control_absolute_position_in_closed_loop_360 := can.Data{0xA4, 0x00, a_speed[0], a_speed[1], a_delt_c_deg[0], a_delt_c_deg[1], a_delt_c_deg[2], a_delt_c_deg[3]} // works

	deci_A_cw := current_in_deci_A_int16(3)
	a_deci_A_cw := deci_A_cw.to_can_data_alley()
	// deci_A_ccw := current_in_deci_A_int16(-12)
	// a_deci_A_ccw := deci_A_ccw.to_can_data_alley()

	data_to_torque_closed_loop_control_cw := can.Data{0xA1, 0x00, 0x00, 0x00, a_deci_A_cw[0], a_deci_A_cw[1], 0x00, 0x00}
	// data_to_torque_closed_loop_control_ccw := can.Data{0xA1, 0x00, 0x00, 0x00, a_deci_A_ccw[0], a_deci_A_ccw[1], 0x00, 0x00}

	s_data_to_position_cw_and_torque_cw_init := []can.Data{
		data_to_control_absolute_position_in_closed_loop_0,
		data_to_torque_closed_loop_control_cw,
	}

	s_data_to_position_cw_and_torque_cw := []can.Data{
		data_to_control_absolute_position_in_closed_loop_360,
		data_to_torque_closed_loop_control_cw,
	}

	s_data_to_torque_cw_and_position_cw := []can.Data{
		data_to_torque_closed_loop_control_cw,
		data_to_control_absolute_position_in_closed_loop_360,
	}

	// data_to_shutdown_motor := can.Data{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	// s_data_to_torque_closed_loop_control := []can.Data{
	// 	data_to_torque_closed_loop_control_cw,
	// 	data_to_torque_closed_loop_control_ccw,
	// }

	// s_data_to_shutdown_motor := []can.Data{
	// 	data_to_shutdown_motor,
	// 	data_to_shutdown_motor,
	// }

	// s_data_to_control_absolute_position_in_closed_loop_360 := []can.Data{
	// 	data_to_control_absolute_position_in_closed_loop_360,
	// 	data_to_control_absolute_position_in_closed_loop_360,
	// }

	slice_of_sequance := [][]can.Data{
		s_data_to_read_multi_turn_motor_position,
		s_data_to_position_cw_and_torque_cw_init,
		s_data_to_position_cw_and_torque_cw,
		s_data_to_torque_cw_and_position_cw,
		// s_data_to_control_absolute_position_in_closed_loop_0,
		// s_data_to_control_absolute_position_in_closed_loop_360,
		// s_data_to_shutdown_motor,
	}

	// data := can.Data{0x30, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	for i, s_data := range slice_of_sequance {
		err = send_command_to_each_motors(ctx, pTx, motor_ids, s_data)
		if err != nil {
			log.Println("error when sending command_to_each_motors...sequence:", i)
			log.Fatal(err)
		}
		log.Println("Waiting for a second ...")
		time.Sleep(3 * time.Second)
	}

	log.Println("wating 1 sec before exiting main() to confirm othe processes will finish properly")
	time.Sleep(time.Second)
	log.Println("Exiting main()")

	// fmt.Printf("Get message: %v\n", msg)
}
