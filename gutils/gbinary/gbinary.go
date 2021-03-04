package gbinary

/*CRC16 计算CRC16 */
func CRC16(data []byte) uint16 {
	u16RegCrc := uint16(0xffff)
	u8TempReg := uint8(0)
	for i := 0; i < len(data); i++ {
		u16RegCrc = u16RegCrc ^ uint16(data[i])
		for j := 0; j < 8; j++ {
			if (u16RegCrc & 0x0001) > 0 {
				u16RegCrc = u16RegCrc>>1 ^ 0xA001
			} else {
				u16RegCrc >>= 1
			}
		}
	}
	u8TempReg = uint8(u16RegCrc >> 8) //hi 高位
	return uint16(uint16(u8TempReg) | (u16RegCrc << 8))
}
