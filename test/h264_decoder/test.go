package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"

	"encoding/hex"
	"encoding/json"

	"time"

	"github.com/kabukky/codec"
	"github.com/nareix/joy4/codec/h264parser"
	flv "github.com/zhangpeihao/goflv"
)

var (
	flvName     = flag.String("flv", "", "flv file with test video")
	frameCount  = flag.Int("frame-count", 48, "frame count for decode")
	outPath     = flag.String("out", "./out", "jpeg out path")
	imgModeFlag = flag.String("img-mode", "copy", "copy or bind")
)

const (
	IMG_MODE_COPY = 1
	IMG_MODE_BIND = 2
)

func main() {
	// process comand line flags
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	codec.SetLogLevel(codec.AV_LOG_INFO)

	// check flv file is set
	if *flvName == "" {
		flag.PrintDefaults()

		return
	}

	var imgMode int
	switch *imgModeFlag {
	case "copy":
		imgMode = IMG_MODE_COPY
	case "bind":
		imgMode = IMG_MODE_COPY
	default:
		flag.PrintDefaults()
		return
	}

	flvFile, err := flv.OpenFile(*flvName)
	if err != nil {
		log.Fatal(err)
	}

	var dec *codec.H264Decoder
	var avcc h264parser.CodecData
	var img *image.YCbCr

	avPacket := codec.NewAVPacket()

	avFrame := codec.NewAVFrame()
	defer avFrame.Release()

	opt := jpeg.Options{Quality: 90}
	btm := time.Time{}

	df := func() {
		switch imgMode {
		case IMG_MODE_COPY:
			// copy frame to img
			err = avFrame.ImgCopy(img)
			if err != nil {
				log.Fatal(err)
			}
		case IMG_MODE_BIND:
			// bind img
			err = avFrame.ImgBind(img)
			if err != nil {
				log.Fatal(err)
			}
		}

		//create out file
		fh, err := os.Create(fmt.Sprintf("%s/frame_%03d.jpg", *outPath, *frameCount))
		if err != nil {
			log.Fatal(err)
		}

		//encode jpeg
		err = jpeg.Encode(fh, img, &opt)
		if err != nil {
			log.Fatal(err)
		}
		fh.Close()
	}

	for !flvFile.IsFinished() {
		header, data, err := flvFile.ReadTag()
		if err != nil {
			log.Fatalln("flvFile.ReadTag() error:", err)
		}

		// skip non video frames
		if header.TagType != flv.VIDEO_TAG {
			continue
		}

		// decoder configuration tag
		if data[0] == 0x17 && data[1] == 0x00 {
			log.Println("Got decoder configuration tag, create decoder")

			fmt.Println(hex.Dump(data[5:]))
			avcc, err = h264parser.NewCodecDataFromAVCDecoderConfRecord(data[5:])
			if err != nil {
				log.Fatal(err)
			}

			b, _ := json.MarshalIndent(avcc, "", "  ")
			fmt.Println(string(b))

			if dec != nil {
				log.Println("decoder created, skip")

				continue
			}

			// create image
			img = image.NewYCbCr(
				image.Rectangle{
					Min: image.Point{0, 0},
					Max: image.Point{int(avcc.SPSInfo.Width), int(avcc.SPSInfo.Height)},
				},
				image.YCbCrSubsampleRatio420,
			)

			// create decoder
			dec, err = codec.NewH264Decoder(data[5:])
			if err != nil {
				log.Fatal(err)
			}
			defer dec.Release()

			continue
		}

		if btm.IsZero() {
			btm = time.Now()
		}

		// set packet data
		avPacket.Data = data[5:]
		avPacket.Pts = int64(header.Timestamp)
		avPacket.Dts = int64(header.Timestamp)

		//log.Printf("Send packet with pts:%d, dts:%d", avPacket.Pts, avPacket.Dts)
		// decode video frame
		got, err := dec.Decode2(avPacket, avFrame)
		if err != nil {
			log.Fatal(err)
		}
		if !got {
			log.Println("Picture delayed")
			continue
		}
		//log.Printf("Decoded frame pts:%d, dts:%d", avFrame.GetPktPts(), avFrame.GetPktDts())

		df()

		*frameCount--
		if *frameCount <= 0 {
			break
		}
	}

	// flush decoder buffer

	avPacket.Data = nil
	avPacket.Pts = 0
	avPacket.Dts = 0
	for {
		// decode video frame
		got, err := dec.Decode2(avPacket, avFrame)
		if err != nil {
			log.Fatal(err)
		}
		if !got {
			log.Println("flush, got == false, exit")
			break
		}
		log.Printf("Decoded frame pts:%d, dts:%d", avFrame.GetPktPts(), avFrame.GetPktDts())

		df()
	}

	diff := time.Now().Sub(btm)
	log.Println("Time elapsed:", diff)
}
