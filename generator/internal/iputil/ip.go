package iputil

import "net/netip"

func Number(a netip.Addr) uint32 {
	b := a.As4()
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func End(p netip.Prefix) uint32 {
	return uint32(uint64(Number(p.Addr())) + (uint64(1) << uint(32-p.Bits())) - 1)
}

func LastAddress(prefix netip.Prefix) netip.Addr {
	b := prefix.Masked().Addr().As16()
	for bit := prefix.Bits(); bit < 128; bit++ {
		b[bit/8] |= 1 << uint(7-bit%8)
	}
	return netip.AddrFrom16(b)
}
