package gui

import (
	"net"
	"testing"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gopkg.in/ini.v1"
)

const testConf = `[Interface]
Address = 10.66.66.1/24
ListenPort = 50709
PrivateKey = 1tGbVvkrT9XCe2IMKfn8TxQuh8dMlwQpcgSyRfkm3vo=

### Peer A
[Peer]
PublicKey = vxejexxPgqffeXFsEnVwVkd7fpSSof/G9vDU/tQT4fM=
AllowedIPs = 10.66.66.2/32

### Peer B
[Peer]
PublicKey = notQsdHNCmKfpKeCiyg99Ybafxyrn/w8bc46Hvz174g=
AllowedIPs = 10.66.66.3/32

### Peer C
[Peer]
PublicKey = Ma2smFQ3mk6syScZhFmITcJIMCHmu9lPI6uvGUF19VM=
AllowedIPs = 10.66.66.4/32
`

func testConfig(t *testing.T) *ini.File {
	t.Helper()

	cfg, err := ini.LoadSources(ini.LoadOptions{AllowNonUniqueSections: true}, []byte(testConf))
	if err != nil {
		t.Fatalf("error loading test config: %v", err)
	}

	return cfg
}

func TestPeerDisplaysEveryConfiguredPeer(t *testing.T) {
	peers := peerDisplays(testConfig(t), nil)

	if len(peers) != 3 {
		t.Fatalf("expected 3 peers, got %d", len(peers))
	}

	want := []string{"10.66.66.2/32", "10.66.66.3/32", "10.66.66.4/32"}
	for i, p := range peers {
		if p.allowedIPs != want[i] {
			t.Errorf("peer %d: expected AllowedIPs %q, got %q", i, want[i], p.allowedIPs)
		}
		if p.handshake != "unknown" || p.transfer != "unknown" {
			t.Errorf("peer %d: expected unknown stats without a live device", i)
		}
		if p.active {
			t.Errorf("peer %d: expected inactive without a live device", i)
		}
	}

	if peers[0].comment != "### Peer A" {
		t.Errorf("unexpected comment for peer 0: %q", peers[0].comment)
	}
}

func TestPeerDisplaysMergesLiveDeviceData(t *testing.T) {
	key, err := wgtypes.ParseKey("notQsdHNCmKfpKeCiyg99Ybafxyrn/w8bc46Hvz174g=")
	if err != nil {
		t.Fatal(err)
	}

	_, ipnet, err := net.ParseCIDR("10.66.66.3/32")
	if err != nil {
		t.Fatal(err)
	}

	device := &wgtypes.Device{
		Name: "wg0",
		Peers: []wgtypes.Peer{{
			PublicKey:         key,
			Endpoint:          &net.UDPAddr{IP: net.ParseIP("192.168.1.50"), Port: 51820},
			AllowedIPs:        []net.IPNet{*ipnet},
			LastHandshakeTime: time.Now().Add(-90 * time.Second),
			ReceiveBytes:      2048,
			TransmitBytes:     1024,
		}},
	}

	peers := peerDisplays(testConfig(t), device)

	if len(peers) != 3 {
		t.Fatalf("expected 3 peers, got %d", len(peers))
	}

	if peers[1].endpoint != "192.168.1.50:51820" {
		t.Errorf("expected live endpoint for peer 1, got %q", peers[1].endpoint)
	}
	if peers[1].transfer != "2.0 kB received, 1.0 kB sent" {
		t.Errorf("unexpected transfer for peer 1: %q", peers[1].transfer)
	}
	if !peers[1].active {
		t.Errorf("peer 1 should be active (present on the live device)")
	}
	if peers[0].active || peers[2].active {
		t.Errorf("peers without live data should be inactive")
	}

	if peers[0].endpoint != "" || peers[2].endpoint != "" {
		t.Errorf("peers without live data should have no endpoint")
	}
}

func TestPeerDisplaysIncludesExtraLivePeers(t *testing.T) {
	key, err := wgtypes.ParseKey("vxejexxPgqffeXFsEnVwVkd7fpSSof/G9vDU/tQT4fM=")
	if err != nil {
		t.Fatal(err)
	}
	extra, err := wgtypes.ParseKey("Ma2smFQ3mk6syScZhFmITcJIMCHmu9lPI6uvGUF19VM=")
	if err != nil {
		t.Fatal(err)
	}

	device := &wgtypes.Device{
		Name: "wg0",
		Peers: []wgtypes.Peer{
			{PublicKey: key},
			{PublicKey: extra},
		},
	}

	peers := peerDisplays(testConfig(t), device)

	if len(peers) != 3 {
		t.Fatalf("configured peers must not be duplicated, got %d peers", len(peers))
	}

	if !peers[0].active || peers[1].active || !peers[2].active {
		t.Errorf("unexpected active states: %v %v %v", peers[0].active, peers[1].active, peers[2].active)
	}
}

func TestPeerDisplaysAppendsLiveOnlyPeers(t *testing.T) {
	key, err := wgtypes.ParseKey("vxejexxPgqffeXFsEnVwVkd7fpSSof/G9vDU/tQT4fM=")
	if err != nil {
		t.Fatal(err)
	}
	runtimeOnly, err := wgtypes.ParseKey("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa=")
	if err != nil {
		t.Fatal(err)
	}

	_, ipnet, err := net.ParseCIDR("10.66.66.9/32")
	if err != nil {
		t.Fatal(err)
	}

	device := &wgtypes.Device{
		Name: "wg0",
		Peers: []wgtypes.Peer{
			{PublicKey: key},
			{PublicKey: runtimeOnly, AllowedIPs: []net.IPNet{*ipnet}},
		},
	}

	peers := peerDisplays(testConfig(t), device)

	if len(peers) != 4 {
		t.Fatalf("expected 3 configured + 1 runtime peer, got %d peers", len(peers))
	}
	if !peers[3].active {
		t.Errorf("runtime-only peer should be active")
	}
	if peers[3].allowedIPs != "10.66.66.9/32" {
		t.Errorf("runtime-only peer AllowedIPs = %q", peers[3].allowedIPs)
	}
}
