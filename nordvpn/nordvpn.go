package nordvpn

import (
	"context"
	"errors"
	"fmt"
	childprocess "github.com/NordSecurity/nordvpn-linux/child_process"
	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/norduser/process"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
)

var (
	DaemonURL    = fmt.Sprintf("%s://%s", "unix", "/run/nordvpn/nordvpnd.sock")
	DaemonClient pb.DaemonClient
	UserManager  childprocess.ChildProcessManager
)

func getNorduserManager() childprocess.ChildProcessManager {
	if snapconf.IsUnderSnap() {
		usr, err := user.Current()
		if err != nil {
			os.Exit(int(childprocess.CodeFailedToEnable))
		}

		uid, err := strconv.Atoi(usr.Uid)
		if err != nil {
			log.Printf("Invalid unix user id, failed to convert from string: %s", usr.Uid)
			os.Exit(int(childprocess.CodeFailedToEnable))
		}

		return process.NewNorduserGRPCProcessManager(uint32(uid))
	}

	return childprocess.NoopChildProcessManager{}
}

func Init() {
	conn, _ := grpc.Dial(
		DaemonURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	DaemonClient = pb.NewDaemonClient(conn)
	UserManager = getNorduserManager()
}

type Call struct {
	ctx context.Context
}

func PrepareCall() Call {
	UserManager.StartProcess()
	return Call{
		ctx: context.Background(),
	}
}

func (call Call) EndCall() {
	UserManager.StopProcess(false)
}

func (call Call) GetAccount() (*pb.AccountResponse, error) {
	payload, err := DaemonClient.AccountInfo(call.ctx, &pb.AccountRequest{Full: true})
	if err != nil {
		return nil, err
	}

	switch payload.Type {
	case CodeUnauthorized:
		return nil, errors.New("Unauthorized")
	case CodeExpiredRenewToken:
		return nil, errors.New("Got unexptected expired token")
	case CodeTokenRenewError:

		return nil, errors.New("Failed to renew token")
	}

	return payload, nil
}

func (call Call) GetLoginURL() (string, error) {
	resp, err := DaemonClient.LoginOAuth2(
		call.ctx,
		&pb.LoginOAuth2Request{
			Type: pb.LoginType_LoginType_LOGIN,
		},
	)
	if err != nil {
		return "", err
	}

	switch resp.Status {
	case pb.LoginOAuth2Status_UNKNOWN_OAUTH2_ERROR:
		return "", errors.New("Unhandled error")
	case pb.LoginOAuth2Status_NO_NET:
		return "", errors.New("No internet connection")
	case pb.LoginOAuth2Status_ALREADY_LOGGED_IN:
		return "", errors.New("Already logged in")
	case pb.LoginOAuth2Status_SUCCESS:
		if url := resp.Url; url != "" {
			return url, nil
		} else {
			return "", errors.New("Failed to get login url")
		}
	}
	return "", errors.New("Unknown error")
}

func (call Call) GetCountriesIntern() (*pb.ServerGroupsList, error) {
	resp, err := DaemonClient.Countries(call.ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}
	if resp.Type != CodeSuccess {
		return nil, errors.New("Failed to get countries")
	}
	return resp, nil
}

func (call Call) Connect(country string, city string, group string) (string, error) {
	server_tag := country
	if city != "" {
		server_tag += " " + city
	}

	resp, err := DaemonClient.Connect(call.ctx, &pb.ConnectRequest{
		ServerTag:   server_tag,
		ServerGroup: group,
	})
	if err != nil {
		return "", err
	}

	var rpcErr error = nil
	for {
		out, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		switch out.Type {
		case CodeFailure:
			rpcErr = errors.New(client.ConnectCantConnect)
		case CodeServerUnavailable:
			rpcErr = errors.New("Server unavailable")
		case CodeVPNRunning:
			rpcErr = errors.New("Already connected")
		case CodeNothingToDo:
			rpcErr = errors.New("Connection already in progress")
		default:
			return strings.Join(out.Data, "|"), nil
		}
	}
	return "", rpcErr
}

func (call Call) Disconnect() (string, error) {
	resp, err := DaemonClient.Disconnect(call.ctx, &pb.Empty{})

	if err != nil {
		return "false", err
	}

	for {
		out, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "false", err
		}

		switch out.Type {
		case CodeVPNNotRunning:
			return "false", errors.New("Not connected to VPN")
		case CodeDisconnected:
			return "true", nil
		}
	}
	return "false", errors.New("Unknown error")
}

func (call Call) Register() (string, error) {
	resp, err := DaemonClient.LoginOAuth2(
		call.ctx,
		&pb.LoginOAuth2Request{
			Type: pb.LoginType_LoginType_SIGNUP,
		},
	)
	if err != nil {
		return "", err
	}

	switch resp.Status {
	case pb.LoginOAuth2Status_UNKNOWN_OAUTH2_ERROR:
		return "", errors.New("Unhandled error")
	case pb.LoginOAuth2Status_NO_NET:
		return "", errors.New("No internet connection")
	case pb.LoginOAuth2Status_ALREADY_LOGGED_IN:
		return "", errors.New("Already logged in")
	case pb.LoginOAuth2Status_SUCCESS:
		if url := resp.Url; url != "" {
			return url, nil
		} else {
			return "", errors.New("Failed to get login url")
		}
	}
	return "", errors.New("Unknown error")
}

func (call Call) GetStatus() (*pb.StatusResponse, error) {
	resp, err := DaemonClient.Status(call.ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (call Call) LoginWithToken(token string) (string, error) {
	if token == "" {
		return "", errors.New("No token provided")
	}

	resp, err := DaemonClient.LoginWithToken(call.ctx, &pb.LoginWithTokenRequest{
		Token: token,
	})
	if err != nil {
		return "", err
	}
	switch resp.Type {
	case CodeGatewayError:
		return "", errors.New("Gateway error")
	case CodeUnauthorized:
		return "", errors.New("Unauthorized")
	case CodeBadRequest:
		return "", errors.New("Bad request")
	case CodeTokenLoginFailure:
		return "", errors.New("Token login failure")
	case CodeTokenInvalid:
		return "", errors.New("Token invalid")
	}
	return "true", nil
}
