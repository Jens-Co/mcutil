package description

// Color is a Minecraft color code type
type Color rune

var (
	// Black is Minecraft color code §0 (#000000)
	Black Color = '0'
	// DarkBlue is Minecraft color code §1 (#0000aa)
	DarkBlue Color = '1'
	// DarkGreen is Minecraft color code §2 (#00aa00)
	DarkGreen Color = '2'
	// DarkAqua is Minecraft color code §3 (#00aaaa)
	DarkAqua Color = '3'
	// DarkRed is Minecraft color code §4 (#aa0000)
	DarkRed Color = '4'
	// DarkPurple is Minecraft color code §5 (#aa00aa)
	DarkPurple Color = '5'
	// Gold is Minecraft color code §6 (#ffaa00)
	Gold Color = '6'
	// Gray is Minecraft color code §7 (#aaaaaa)
	Gray Color = '7'
	// DarkGray is Minecraft color code §8 (#555555)
	DarkGray Color = '8'
	// Blue is Minecraft color code §9 (#5555ff)
	Blue Color = '9'
	// Green is Minecraft color code §a (#55ff55)
	Green Color = 'a'
	// Aqua is Minecraft color code §b (#55ffff)
	Aqua Color = 'b'
	// Red is Minecraft color code §c (#ff5555)
	Red Color = 'c'
	// LightPurple is Minecraft color code §d (#ff55ff)
	LightPurple Color = 'd'
	// Yellow is Minecraft color code §e (#ffff55)
	Yellow Color = 'e'
	// White is Minecraft color code §f (#ffffff)
	White Color = 'f'
	// MinecoinGold is Minecraft color code §g (#ddd605)
	MinecoinGold Color = 'g'
)

// ParseColor attempts to return a Color type based on a color code string, color name string, or a Color type itself
func ParseColor(value interface{}) Color {
	switch value {
	case "0", "black", Black:
		return Black
	case "1", "dark_blue", DarkBlue:
		return DarkBlue
	case "2", "dark_green", DarkGreen:
		return DarkGreen
	case "3", "dark_aqua", DarkAqua:
		return DarkAqua
	case "4", "dark_red", DarkRed:
		return DarkRed
	case "5", "dark_purple", DarkPurple:
		return DarkPurple
	case "6", "gold", Gold:
		return Gold
	case "7", "gray", Gray:
		return Gray
	case "8", "dark_gray", DarkGray:
		return DarkGray
	case "9", "blue", Blue:
		return Blue
	case "a", "green", Green:
		return Green
	case "b", "aqua", Aqua:
		return Aqua
	case "c", "red", Red:
		return Red
	case "d", "light_purple", LightPurple:
		return LightPurple
	case "e", "yellow", Yellow:
		return Yellow
	case "f", "white", White:
		return White
	case "g", "minecoin_gold", MinecoinGold:
		return MinecoinGold
	default:
		return White
	}
}

// ToRaw returns the encoded Minecraft formatting of the color (§ + code)
func (c Color) ToRaw() string {
	return "\u00A7" + string(c)
}

// ToHex returns the hex string of the color prefixed with a # symbol
func (c Color) ToHex() string {
	switch c {
	case Black:
		return "#000000"
	case DarkBlue:
		return "#0000aa"
	case DarkGreen:
		return "#00aa00"
	case DarkAqua:
		return "#00aaaa"
	case DarkRed:
		return "#aa0000"
	case DarkPurple:
		return "#aa00aa"
	case Gold:
		return "#ffaa00"
	case Gray:
		return "#aaaaaa"
	case DarkGray:
		return "#555555"
	case Blue:
		return "#5555ff"
	case Green:
		return "#55ff55"
	case Aqua:
		return "#55ffff"
	case Red:
		return "#ff5555"
	case LightPurple:
		return "#ff55ff"
	case Yellow:
		return "#ffff55"
	case White:
		return "#ffffff"
	case MinecoinGold:
		return "#ddd605"
	default:
		return "#ffffff"
	}
}
