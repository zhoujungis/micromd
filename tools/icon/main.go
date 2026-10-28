//go:build ignore

package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
)

type rgb struct{ r, g, b float64 }

var (
	topColor    = rgb{0.31, 0.53, 1.0}
	bottomColor = rgb{0.11, 0.32, 0.80}
	markColor   = rgb{1, 1, 1}
)

func main() {
	outDir := "assets"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		panic(err)
	}
	must(os.WriteFile(filepath.Join(outDir, "icon.png"), encodePNG(256), 0o644))
	must(os.WriteFile(filepath.Join(outDir, "icon.ico"), buildICO([]int{16, 24, 32, 48, 64, 128, 256}), 0o644))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func encodePNG(size int) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, shade(size)); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func buildICO(sizes []int) []byte {
	sort.Ints(sizes)
	images := make([][]byte, len(sizes))
	for i, size := range sizes {
		images[i] = encodePNG(size)
	}

	var buf bytes.Buffer
	writeU16(&buf, 0)
	writeU16(&buf, 1)
	writeU16(&buf, uint16(len(sizes)))
	offset := 6 + 16*len(sizes)
	for i, size := range sizes {
		dim := byte(size)
		if size >= 256 {
			dim = 0
		}
		buf.Write([]byte{dim, dim, 0, 0})
		writeU16(&buf, 1)
		writeU16(&buf, 32)
		writeU32(&buf, uint32(len(images[i])))
		writeU32(&buf, uint32(offset))
		offset += len(images[i])
	}
	for _, data := range images {
		buf.Write(data)
	}
	return buf.Bytes()
}

func writeU16(buf *bytes.Buffer, v uint16) { _ = binary.Write(buf, binary.LittleEndian, v) }
func writeU32(buf *bytes.Buffer, v uint32) { _ = binary.Write(buf, binary.LittleEndian, v) }

func shade(size int) *image.NRGBA {
	const samples = 4
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	half := float64(size) / 2
	margin := math.Max(1, float64(size)*0.045)
	radius := float64(size) * 0.225
	hw := half - margin
	hh := half - margin
	stroke := math.Max(1.6, float64(size)*0.092)

	mark := [][][2]float64{
		{{0.16, 0.70}, {0.16, 0.31}, {0.36, 0.53}, {0.56, 0.31}, {0.56, 0.70}},
		{{0.79, 0.29}, {0.79, 0.66}},
		{{0.70, 0.57}, {0.79, 0.66}, {0.88, 0.57}},
	}
	for i := range mark {
		for j := range mark[i] {
			mark[i][j][0] *= float64(size)
			mark[i][j][1] *= float64(size)
		}
	}

	total := float64(samples * samples)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			var pr, pg, pb, pa float64
			for sy := 0; sy < samples; sy++ {
				for sx := 0; sx < samples; sx++ {
					px := float64(x) + (float64(sx)+0.5)/samples
					py := float64(y) + (float64(sy)+0.5)/samples

					dist := sdRoundRect(px, py, half, half, hw, hh, radius)
					coverage := clamp01(0.5 - dist)
					if coverage <= 0 {
						continue
					}

					t := clamp01((py - (half - hh)) / (2 * hh))
					col := mix(topColor, bottomColor, t)

					ink := 0.0
					for _, line := range mark {
						nearest := math.Inf(1)
						for i := 0; i < len(line)-1; i++ {
							nearest = math.Min(nearest, sdSegment(px, py, line[i][0], line[i][1], line[i+1][0], line[i+1][1]))
						}
						ink = math.Max(ink, clamp01(0.5-(nearest-stroke/2)))
					}
					if ink > 0 {
						col = mix(col, markColor, ink)
					}

					pr += col.r * coverage
					pg += col.g * coverage
					pb += col.b * coverage
					pa += coverage
				}
			}
			if pa == 0 {
				continue
			}
			alpha := pa / total
			img.Set(x, y, color.NRGBA{
				R: clamp8(pr / pa * 255),
				G: clamp8(pg / pa * 255),
				B: clamp8(pb / pa * 255),
				A: clamp8(alpha * 255),
			})
		}
	}
	return img
}

func sdRoundRect(px, py, cx, cy, hw, hh, r float64) float64 {
	qx := math.Abs(px-cx) - (hw - r)
	qy := math.Abs(py-cy) - (hh - r)
	return math.Hypot(math.Max(qx, 0), math.Max(qy, 0)) + math.Min(math.Max(qx, qy), 0) - r
}

func sdSegment(px, py, ax, ay, bx, by float64) float64 {
	abx, aby := bx-ax, by-ay
	denom := abx*abx + aby*aby
	t := 0.0
	if denom > 0 {
		t = ((px-ax)*abx + (py-ay)*aby) / denom
		t = clamp01(t)
	}
	return math.Hypot(px-(ax+abx*t), py-(ay+aby*t))
}

func mix(a, b rgb, t float64) rgb {
	return rgb{a.r + (b.r-a.r)*t, a.g + (b.g-a.g)*t, a.b + (b.b-a.b)*t}
}

func clamp01(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}

func clamp8(v float64) uint8 {
	return uint8(math.Round(math.Max(0, math.Min(255, v))))
}
