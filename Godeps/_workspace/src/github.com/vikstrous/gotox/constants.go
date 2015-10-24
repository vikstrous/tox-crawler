package gotox

const (
	UserStatusNone = iota
	UserStatusAway
	UserStatusBusy
)

const PingIDSize = 8
const SendbackDataSize = 8
const NonceSize = 24
const SymmetricKeySize = 32
const PublicKeySize = 32
const PrivateKeySize = 32
const MaxCloseClients = 32

// Tox addresses are in the format
// * [Public Key (TOX_PUBLIC_KEY_SIZE bytes)][nospam (4 bytes)][checksum (2 bytes)].
const AddressSize = PublicKeySize + 4 + 2
const MaxNameLength = 128
const MaxStatusMessageLength = 1007
const MaxFriendRequestLength = 1016
const MaxMessageLength = 1372
const MaxCustomPacketLength = 1373
const HashLength = 32
const FileIDLength = 32
const MaxFilenameLength = 255
