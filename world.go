package gokart

import "math"

type World struct {
	Tracks []*Track
}

func (w World) GetTrack(points []Timely) (t *Track) {
	minD := -1.0
	for _, tr := range w.Tracks {
		for _, pt := range points {
			gps := pt.Value.(GPS5)
			if gps.Accuracy >= 10000 {
				// not precise enough
				continue
			}
			if minD < 0 || math.Abs(tr.To(gps)) < minD {
				minD = math.Abs(tr.To(gps))
				t = tr
			}
		}
	}
	return
}

var (
	Cik = Track{
		Name:     "CIK Le Mans",
		LogoFile: "../../images/le-mans-karting-logo.png",
		LogoMask: "../../images/le-mans-karting-logo-mask.png",
		Start: NewLine(
			47.94267956470532, 0.21127426211668823, // real start line
			47.94268450574376, 0.21142312470907643), // real start line
		//NewLine(
		//	47.94251032066459, 0.21129713002829278, // Alfano
		//	47.942509630320565, 0.21143213273473976), // Alfano
		Sectors: []Line{
			NewLine(
				47.94251613207653, 0.21202198640034764,
				47.94252829837948, 0.2121507716882105),
			NewLine(
				47.942641567263436, 0.21328805762477007,
				47.94263718903707, 0.21339698916760877),
		},
		Limits: NewLine(47.9411527, 0.211025, 47.943917, 0.2136273),
	}
	/*	Itekkarting = Track{
			"ItekKarting",
			"../../images/itekkarting-logo.png",
			"",
			// 45.69072118184242, 0.18885285434343446, // real line
			// 45.690667315626975, 0.18890247520701825, // real line
			NewLine(
				45.69029938660369, 0.18865154225196978, // transponder
				45.690308693450284, 0.1887526429814874), // transponder
			[]Line{},
		}
	*/
	// Itekkarting, sectors for short version
	Itekkarting = Track{
		Name:     "ItekKarting",
		LogoFile: "../../images/itekkarting-logo.png",
		Start: NewLine(
			45.69072118184242, 0.18885285434343446, // real line
			45.690667315626975, 0.18890247520701825), // real line Silverstone
		Sectors: []Line{
			NewLine(
				45.69092581023422, 0.18972201698773333,
				45.69087384069635, 0.1897738134418096), // Monza la Parabolique
			NewLine(
				45.69038571938242, 0.18951671467106518,
				45.69032256591012, 0.1895703946325624), // Interlago 'S' de Senna
			NewLine(
				45.6906475423936, 0.18897614802442725,
				45.69059030983027, 0.18904677955271307), // Laguna Seca La Crosse
			NewLine(
				45.690214020708886, 0.18900440063448615,
				45.69025151816443, 0.18909669249811292), // Formula Kart Speedway
			NewLine(
				45.689967326290514, 0.18942630630069726,
				45.68990154092989, 0.18948092801590494), // Variante (Short)
			NewLine(
				45.69008113478424, 0.18885372002735068,
				45.69012521084291, 0.18894601189097746), // Indianapolis
		},
		Limits: NewLine(
			45.68967867893401, 0.18834913948058676,
			45.691401470702985, 0.1909805320719727),
	}
	Ancenis = Track{
		Name: "Ancenis",
		Start: NewLine(
			47.39503548385757, -1.1856938139621283, // real line
			47.3949784858503, -1.1856306643524066), // real line
		Sectors: []Line{
			NewLine(
				47.396066376082544, -1.1859248457192473,
				47.39607454654443, -1.185835662274258),
			NewLine(
				47.39623290444295, -1.1846790469523996,
				47.39616935331883, -1.1847141342033636),
		},
		Limits: NewLine(47.3947574, -1.1866685, 47.3968351, -1.1841471),
	}
	ValDeVienne = Track{
		Name:     "Val de Vienne",
		LogoFile: "../../images/val-de-vienne-logo.png",
		LogoMask: "../../images/val-de-vienne-logo-mask.png",
		Start: NewLine(
			46.19765536378354, 0.6338169652380998, // Real line
			46.19751148082728, 0.633764662160498), // Real line
		Sectors: []Line{
			NewLine(
				46.194437786686095, 0.6340328808333912,
				46.19427563630064, 0.6339554596033226),
			NewLine(
				46.19724933921958, 0.6294288265480215,
				46.19710812613646, 0.6295336662909374),
		},
		Limits: NewLine(46.1937425, 0.623798, 46.1981878, 0.64001),
	}
	TheWorld = World{
		Tracks: []*Track{&Cik, &Itekkarting, &Ancenis, &ValDeVienne},
	}
)

func ExtractLimits(gps []Timely) (limits Line) {
	for _, p := range gps {
		v := p.Value.(GPS5)
		if limits.P1.Latitude == 0. || v.Latitude < limits.P1.Latitude {
			limits.P1.Latitude = v.Latitude
		}
		if limits.P1.Longitude == 0. || v.Longitude < limits.P1.Longitude {
			limits.P1.Longitude = v.Longitude
		}
		if limits.P2.Latitude == 0. || v.Latitude > limits.P2.Latitude {
			limits.P2.Latitude = v.Latitude
		}
		if limits.P2.Longitude == 0. || v.Longitude > limits.P2.Longitude {
			limits.P2.Longitude = v.Longitude
		}
	}
	return
}
