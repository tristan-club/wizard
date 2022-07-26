package tstore_pb

func (s *Suit) CombatPower() int32 {
	mainScore := (s.Engine.Lv + s.Core.Lv + s.Weapon.Lv + s.Accessory.Lv + s.Chest.Lv - 6) * 60
	minorScore := (s.Leg.Lv + s.Trunk.Lv - 2) * 30
	return 4000 + mainScore + minorScore
}

func (s *Suit) Level() int32 {
	min := s.Trunk.Lv
	if min < s.Leg.Lv {
		min = s.Leg.Lv
	}
	if min < s.Chest.Lv {
		min = s.Chest.Lv
	}
	if min < s.Accessory.Lv {
		min = s.Accessory.Lv
	}
	if min < s.Weapon.Lv {
		min = s.Weapon.Lv
	}
	if min < s.Core.Lv {
		min = s.Core.Lv
	}
	if min < s.Engine.Lv {
		min = s.Engine.Lv
	}
	if min < s.Arm.Lv {
		min = s.Arm.Lv
	}
	return min
}

func (u *User) GetInfo() map[string]string {
	return map[string]string{
		"discord":  u.DiscordName,
		"telegram": u.TeleName,
		"address":  u.Address,
	}
}
