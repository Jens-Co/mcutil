package mcutil

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/mcstatus-io/mcutil/description"
	"github.com/mcstatus-io/mcutil/options"
	"github.com/mcstatus-io/mcutil/response"
)

var (
	defaultJavaStatusOptions = options.JavaStatus{
		EnableSRV:        true,
		Timeout:          time.Second * 5,
		ProtocolVersion:  47,
		DefaultMOTDColor: description.White,
	}
)

type rawJavaStatus struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int64  `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    *int64 `json:"max"`
		Online *int64 `json:"online"`
		Sample []struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"sample"`
	} `json:"players"`
	Description interface{} `json:"description"`
	Favicon     *string     `json:"favicon"`
	ModInfo     struct {
		List []struct {
			ID      string `json:"modid"`
			Version string `json:"version"`
		} `json:"modList"`
		Type string `json:"type"`
	} `json:"modinfo"`
	ForgeData struct {
		Channels []struct {
			Required bool   `json:"required"`
			Res      string `json:"res"`
			Version  string `json:"version"`
		} `json:"channels"`
		FMLNetworkVersion int `json:"fmlNetworkVersion"`
		Mods              []struct {
			ID      string `json:"modId"`
			Version string `json:"modmarker"`
		} `json:"mods"`
	} `json:"forgeData"`
}

// Status retrieves the status of any 1.7+ Minecraft server
func Status(host string, port uint16, options ...options.JavaStatus) (*response.JavaStatus, error) {
	opts := parseJavaStatusOptions(options...)

	var srvRecord *response.SRVRecord = nil

	if opts.EnableSRV && port == 25565 {
		record, err := LookupSRV("tcp", host, port)

		if err == nil && record != nil {
			host = record.Target
			port = record.Port

			srvRecord = &response.SRVRecord{
				Host: record.Target,
				Port: record.Port,
			}
		}
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), opts.Timeout)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	if err = conn.SetDeadline(time.Now().Add(opts.Timeout)); err != nil {
		return nil, err
	}

	if err = writeJavaStatusHandshakeRequestPacket(conn, int32(opts.ProtocolVersion), host, port); err != nil {
		return nil, err
	}

	if err = writeJavaStatusStatusRequestPacket(conn); err != nil {
		return nil, err
	}

	var serverResponse rawJavaStatus

	if err = readJavaStatusStatusResponsePacket(conn, &serverResponse); err != nil {
		return nil, err
	}

	payload := rand.Int63()

	if err = writeJavaStatusPingPacket(conn, payload); err != nil {
		return nil, err
	}

	pingStart := time.Now()

	if err = readJavaStatusPongPacket(conn, payload); err != nil {
		return nil, err
	}

	latency := time.Since(pingStart)

	return formatJavaStatusResponse(serverResponse, srvRecord, latency, opts)
}

func parseJavaStatusOptions(opts ...options.JavaStatus) options.JavaStatus {
	if len(opts) < 1 {
		return defaultJavaStatusOptions
	}

	return opts[0]
}

// https://wiki.vg/Server_List_Ping#Handshake
func writeJavaStatusHandshakeRequestPacket(w io.Writer, protocolVersion int32, host string, port uint16) error {
	buf := &bytes.Buffer{}

	// Packet ID - varint
	if err := writeVarInt(0x00, buf); err != nil {
		return err
	}

	// Protocol version - varint
	if err := writeVarInt(protocolVersion, buf); err != nil {
		return err
	}

	// Host - string
	if err := writeString(host, buf); err != nil {
		return err
	}

	// Port - uint16
	if err := binary.Write(buf, binary.BigEndian, port); err != nil {
		return err
	}

	// Next state - varint
	if err := writeVarInt(1, buf); err != nil {
		return err
	}

	return writePacket(w, buf)
}

// https://wiki.vg/Server_List_Ping#Request
func writeJavaStatusStatusRequestPacket(w io.Writer) error {
	buf := &bytes.Buffer{}

	// Packet ID - varint
	if err := writeVarInt(0x00, buf); err != nil {
		return err
	}

	return writePacket(w, buf)
}

// https://wiki.vg/Server_List_Ping#Response
func readJavaStatusStatusResponsePacket(r io.Reader, result interface{}) error {
	// Packet length - varint
	{
		if _, err := readVarInt(r); err != nil {
			return err
		}
	}

	// Packet type - varint
	{
		packetType, err := readVarInt(r)

		if err != nil {
			return err
		}

		if packetType != 0x00 {
			return fmt.Errorf("status: received unexpected packet type (expected=0x00, received=0x%02X)", packetType)
		}
	}

	// Data - string
	{
		data, err := readString(r)

		if err != nil {
			return err
		}

		if err = json.Unmarshal(data, result); err != nil {
			return err
		}
	}

	return nil
}

// https://wiki.vg/Server_List_Ping#Ping
func writeJavaStatusPingPacket(w io.Writer, payload int64) error {
	buf := &bytes.Buffer{}

	// Packet ID - varint
	if err := writeVarInt(0x01, buf); err != nil {
		return err
	}

	// Payload - int64
	if err := binary.Write(buf, binary.BigEndian, payload); err != nil {
		return err
	}

	return writePacket(w, buf)
}

// https://wiki.vg/Server_List_Ping#Pong
func readJavaStatusPongPacket(r io.Reader, payload int64) error {
	// Packet length - varint
	{
		if _, err := readVarInt(r); err != nil {
			return err
		}
	}

	// Packet type - varint
	{
		packetType, err := readVarInt(r)

		if err != nil {
			return err
		}

		if packetType != 0x01 {
			return fmt.Errorf("status: received unexpected packet type (expected=0x01, received=0x%02X)", packetType)
		}
	}

	// Payload - int64
	{
		var returnPayload int64

		if err := binary.Read(r, binary.BigEndian, &returnPayload); err != nil {
			return err
		}

		if payload != returnPayload {
			return fmt.Errorf("status: received unexpected payload (expected=%X, received=%x)", payload, returnPayload)
		}
	}

	return nil
}

func formatJavaStatusResponse(serverResponse rawJavaStatus, srvRecord *response.SRVRecord, latency time.Duration, opts options.JavaStatus) (*response.JavaStatus, error) {
	motd, err := description.ParseFormatting(serverResponse.Description, opts.DefaultMOTDColor)

	if err != nil {
		return nil, err
	}

	samplePlayers := make([]response.SamplePlayer, 0)

	if serverResponse.Players.Sample != nil {
		for _, player := range serverResponse.Players.Sample {
			name, err := description.ParseFormatting(player.Name)

			if err != nil {
				return nil, err
			}

			samplePlayers = append(samplePlayers, response.SamplePlayer{
				ID:        player.ID,
				NameRaw:   name.Raw,
				NameClean: name.Clean,
				NameHTML:  name.HTML,
			})
		}
	}

	version, err := description.ParseFormatting(serverResponse.Version.Name)

	if err != nil {
		return nil, err
	}

	result := &response.JavaStatus{
		Version: response.Version{
			NameRaw:   version.Raw,
			NameClean: version.Clean,
			NameHTML:  version.HTML,
			Protocol:  serverResponse.Version.Protocol,
		},
		Players: response.Players{
			Online: serverResponse.Players.Online,
			Max:    serverResponse.Players.Max,
			Sample: samplePlayers,
		},
		MOTD:      *motd,
		Favicon:   serverResponse.Favicon,
		SRVResult: srvRecord,
		Latency:   latency,
		ModInfo:   nil,
	}

	if len(serverResponse.ModInfo.Type) > 0 {
		mods := make([]response.Mod, 0)

		for _, mod := range serverResponse.ModInfo.List {
			mods = append(mods, response.Mod{
				ID:      mod.ID,
				Version: mod.Version,
			})
		}

		result.ModInfo = &response.ModInfo{
			Type: serverResponse.ModInfo.Type,
			Mods: mods,
		}
	}

	if serverResponse.ForgeData.Mods != nil {
		mods := make([]response.Mod, 0)

		for _, mod := range serverResponse.ForgeData.Mods {
			mods = append(mods, response.Mod{
				ID:      mod.ID,
				Version: mod.Version,
			})
		}

		result.ModInfo = &response.ModInfo{
			Type: "FML2",
			Mods: mods,
		}
	}

	return result, nil
}
