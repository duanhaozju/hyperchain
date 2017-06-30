package network_test

import (
	"testing"
	"hyperchain/p2p/network"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"hyperchain/p2p/network/mock"
	pb "hyperchain/p2p/message"
)

func TestClient_Chat(t *testing.T) {
	Convey("client chat",t, func() {
		controller := gomock.NewController(t)
		mock_ChatClient := mock_network.NewMockChatClient(controller)
		chat_chatClient := mock_network.NewMockChat_ChatClient(controller)
		//this addr will not actually connect
		client,err := network.NewClient("localhost:50015")
		So(err,ShouldBeNil)
		Convey("test Greeting", func() {
			mock_ChatClient.EXPECT().Greeting(gomock.Any(),gomock.Any()).Return(&pb.Message{MessageType:pb.MsgType_HELLO_RESPONSE,},nil)
			err := client.Connect(mock_ChatClient)
			So(err,ShouldBeNil)
			resp,err := client.Greeting(&pb.Message{MessageType:pb.MsgType_HELLO,})
			So(err,ShouldBeNil)
			So(resp.MessageType,ShouldEqual,pb.MsgType_HELLO_RESPONSE)
		})
		Convey("test Wisper", func() {
			mock_ChatClient.EXPECT().Wisper(gomock.Any(),gomock.Any()).Return(&pb.Message{MessageType:pb.MsgType_KEEPALIVE,},nil)
			err := client.Connect(mock_ChatClient)
			So(err,ShouldBeNil)
			resp,err := client.Whisper(&pb.Message{MessageType:pb.MsgType_KEEPALIVE,})
			So(err,ShouldBeNil)
			So(resp.MessageType,ShouldEqual,pb.MsgType_KEEPALIVE)
		})

		Convey("Test chat", func() {
			chat_chatClient.EXPECT().Recv().Return(&pb.Message{MessageType:pb.MsgType_HELLO},nil)
			chat_chatClient.EXPECT().Send(gomock.Any()).Return(nil)
			mock_ChatClient.EXPECT().Chat(gomock.Any()).Return(chat_chatClient,nil)
			err := client.Chat()
			So(err,ShouldBeNil)
		})

		Convey("Client close", func() {
			err := client.Close()
			So(err,ShouldBeNil)
		})
	})
}
