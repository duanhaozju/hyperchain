package network_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	pb "hyperchain/p2p/message"
	"hyperchain/p2p/network"
	"hyperchain/p2p/network/mock"
	"testing"
)

func TestClient_Chat(t *testing.T) {
	Convey("client chat", t, func() {
		controller := gomock.NewController(t)
		mock_ChatClient := mock_network.NewMockChatClient(controller)
		chat_chatClient := mock_network.NewMockChat_ChatClient(controller)
		//this addr will not actually connect
		client, err := network.NewClient("node1", "localhost:50015", nil)
		So(err, ShouldBeNil)

		Convey("Test chat", func() {
			chat_chatClient.EXPECT().Recv().Return(&pb.Message{MessageType: pb.MsgType_HELLO}, nil)
			chat_chatClient.EXPECT().Send(gomock.Any()).Return(nil)
			mock_ChatClient.EXPECT().Chat(gomock.Any()).Return(chat_chatClient, nil)
			err := client.Chat()
			So(err, ShouldBeNil)
		})

	})
}
