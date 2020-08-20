package clientpacket

// ctorTable is the table of all packet constructors.
var ctorTable = []func([]byte) Packet{
	/*00*/ nil,
	/*01*/ nil,
	/*02*/ nil,
	/*03*/ nil,
	/*04*/ nil,
	/*05*/ nil,
	/*06*/ nil,
	/*07*/ nil,
	/*08*/ nil,
	/*09*/ nil,
	/*0A*/ nil,
	/*0B*/ nil,
	/*0C*/ nil,
	/*0D*/ nil,
	/*0E*/ nil,
	/*0F*/ nil,
	/*10*/ nil,
	/*11*/ nil,
	/*12*/ nil,
	/*13*/ nil,
	/*14*/ nil,
	/*15*/ nil,
	/*16*/ nil,
	/*17*/ nil,
	/*18*/ nil,
	/*19*/ nil,
	/*1A*/ nil,
	/*1B*/ nil,
	/*1C*/ nil,
	/*1D*/ nil,
	/*1E*/ nil,
	/*1F*/ nil,
	/*20*/ nil,
	/*21*/ nil,
	/*22*/ nil,
	/*23*/ nil,
	/*24*/ nil,
	/*25*/ nil,
	/*26*/ nil,
	/*27*/ nil,
	/*28*/ nil,
	/*29*/ nil,
	/*2A*/ nil,
	/*2B*/ nil,
	/*2C*/ nil,
	/*2D*/ nil,
	/*2E*/ nil,
	/*2F*/ nil,
	/*30*/ nil,
	/*31*/ nil,
	/*32*/ nil,
	/*33*/ nil,
	/*34*/ nil,
	/*35*/ nil,
	/*36*/ nil,
	/*37*/ nil,
	/*38*/ nil,
	/*39*/ nil,
	/*3A*/ nil,
	/*3B*/ nil,
	/*3C*/ nil,
	/*3D*/ nil,
	/*3E*/ nil,
	/*3F*/ nil,
	/*40*/ nil,
	/*41*/ nil,
	/*42*/ nil,
	/*43*/ nil,
	/*44*/ nil,
	/*45*/ nil,
	/*46*/ nil,
	/*47*/ nil,
	/*48*/ nil,
	/*49*/ nil,
	/*4A*/ nil,
	/*4B*/ nil,
	/*4C*/ nil,
	/*4D*/ nil,
	/*4E*/ nil,
	/*4F*/ nil,
	/*50*/ nil,
	/*51*/ nil,
	/*52*/ nil,
	/*53*/ nil,
	/*54*/ nil,
	/*55*/ nil,
	/*56*/ nil,
	/*57*/ nil,
	/*58*/ nil,
	/*59*/ nil,
	/*5A*/ nil,
	/*5B*/ nil,
	/*5C*/ nil,
	/*5D*/ nil,
	/*5E*/ nil,
	/*5F*/ nil,
	/*60*/ nil,
	/*61*/ nil,
	/*62*/ nil,
	/*63*/ nil,
	/*64*/ nil,
	/*65*/ nil,
	/*66*/ nil,
	/*67*/ nil,
	/*68*/ nil,
	/*69*/ nil,
	/*6A*/ nil,
	/*6B*/ nil,
	/*6C*/ nil,
	/*6D*/ nil,
	/*6E*/ nil,
	/*6F*/ nil,
	/*70*/ nil,
	/*71*/ nil,
	/*72*/ nil,
	/*73*/ nil,
	/*74*/ nil,
	/*75*/ nil,
	/*76*/ nil,
	/*77*/ nil,
	/*78*/ nil,
	/*79*/ nil,
	/*7A*/ nil,
	/*7B*/ nil,
	/*7C*/ nil,
	/*7D*/ nil,
	/*7E*/ nil,
	/*7F*/ nil,
	/*80*/ newAccountLogin,
	/*81*/ nil,
	/*82*/ nil,
	/*83*/ nil,
	/*84*/ nil,
	/*85*/ nil,
	/*86*/ nil,
	/*87*/ nil,
	/*88*/ nil,
	/*89*/ nil,
	/*8A*/ nil,
	/*8B*/ nil,
	/*8C*/ nil,
	/*8D*/ nil,
	/*8E*/ nil,
	/*8F*/ nil,
	/*90*/ nil,
	/*91*/ newGameServerLogin,
	/*92*/ nil,
	/*93*/ nil,
	/*94*/ nil,
	/*95*/ nil,
	/*96*/ nil,
	/*97*/ nil,
	/*98*/ nil,
	/*99*/ nil,
	/*9A*/ nil,
	/*9B*/ nil,
	/*9C*/ nil,
	/*9D*/ nil,
	/*9E*/ nil,
	/*9F*/ nil,
	/*A0*/ newSelectServer,
	/*A1*/ nil,
	/*A2*/ nil,
	/*A3*/ nil,
	/*A4*/ nil,
	/*A5*/ nil,
	/*A6*/ nil,
	/*A7*/ nil,
	/*A8*/ nil,
	/*A9*/ nil,
	/*AA*/ nil,
	/*AB*/ nil,
	/*AC*/ nil,
	/*AD*/ nil,
	/*AE*/ nil,
	/*AF*/ nil,
	/*B0*/ nil,
	/*B1*/ nil,
	/*B2*/ nil,
	/*B3*/ nil,
	/*B4*/ nil,
	/*B5*/ nil,
	/*B6*/ nil,
	/*B7*/ nil,
	/*B8*/ nil,
	/*B9*/ nil,
	/*BA*/ nil,
	/*BB*/ nil,
	/*BC*/ nil,
	/*BD*/ nil,
	/*BE*/ nil,
	/*BF*/ nil,
	/*C0*/ nil,
	/*C1*/ nil,
	/*C2*/ nil,
	/*C3*/ nil,
	/*C4*/ nil,
	/*C5*/ nil,
	/*C6*/ nil,
	/*C7*/ nil,
	/*C8*/ nil,
	/*C9*/ nil,
	/*CA*/ nil,
	/*CB*/ nil,
	/*CC*/ nil,
	/*CD*/ nil,
	/*CE*/ nil,
	/*CF*/ nil,
	/*D0*/ nil,
	/*D1*/ nil,
	/*D2*/ nil,
	/*D3*/ nil,
	/*D4*/ nil,
	/*D5*/ nil,
	/*D6*/ nil,
	/*D7*/ nil,
	/*D8*/ nil,
	/*D9*/ nil,
	/*DA*/ nil,
	/*DB*/ nil,
	/*DC*/ nil,
	/*DD*/ nil,
	/*DE*/ nil,
	/*DF*/ nil,
	/*E0*/ nil,
	/*E1*/ nil,
	/*E2*/ nil,
	/*E3*/ nil,
	/*E4*/ nil,
	/*E5*/ nil,
	/*E6*/ nil,
	/*E7*/ nil,
	/*E8*/ nil,
	/*E9*/ nil,
	/*EA*/ nil,
	/*EB*/ nil,
	/*EC*/ nil,
	/*ED*/ nil,
	/*EE*/ nil,
	/*EF*/ nil,
	/*F0*/ nil,
	/*F1*/ nil,
	/*F2*/ nil,
	/*F3*/ nil,
	/*F4*/ nil,
	/*F5*/ nil,
	/*F6*/ nil,
	/*F7*/ nil,
	/*F8*/ nil,
	/*F9*/ nil,
	/*FA*/ nil,
	/*FB*/ nil,
	/*FC*/ nil,
	/*FD*/ nil,
	/*FE*/ nil,
	/*FF*/ nil,
}
