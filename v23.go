package id3

import (
	"hash/crc32"
	"sync"
)

var (
	v23Data     *versionData
	v23DataInit sync.Once
)

type codec23 struct {
	vdata *versionData
}

func newCodec23() *codec23 {
	v23DataInit.Do(func() {
		v23Data = &versionData{
			headerFlags: flagMap{
				{1 << 7, uint32(TagFlagUnsync)},
				{1 << 6, uint32(TagFlagExtended)},
				{1 << 5, uint32(TagFlagExperimental)},
			},
			headerExFlags: flagMap{
				{1 << 15, uint32(TagFlagHasCRC)},
			},
			frameFlags: flagMap{
				{1 << 15, uint32(FrameFlagDiscardOnTagAlteration)},
				{1 << 14, uint32(FrameFlagDiscardOnFileAlteration)},
				{1 << 13, uint32(FrameFlagReadOnly)},
				{1 << 7, uint32(FrameFlagCompressed)},
				{1 << 6, uint32(FrameFlagEncrypted)},
				{1 << 5, uint32(FrameFlagHasGroupID)},
			},
			bounds: boundsMap{
				"Encoding":         {0, 3, ErrInvalidEncoding},
				"EncryptMethod":    {0x80, 0xff, ErrInvalidEncryptMethod},
				"GroupID":          {0x80, 0xff, ErrInvalidGroupID},
				"LyricContentType": {0, 6, ErrInvalidLyricContentType},
				"PictureType":      {0, 20, ErrInvalidPictureType},
				"TimeStampFormat":  {1, 2, ErrInvalidTimeStampFormat},
			},
			frameTypes: newFrameTypeMap(map[FrameType]string{
				FrameTypeAttachedPicture:              "APIC",
				FrameTypeAudioEncryption:              "AENC",
				FrameTypeComment:                      "COMM",
				FrameTypeEncryptionMethodRegistration: "ENCR",
				FrameTypeGroupID:                      "GRID",
				FrameTypePlayCount:                    "PCNT",
				FrameTypePopularimeter:                "POPM",
				FrameTypePrivate:                      "PRIV",
				FrameTypeLyricsSync:                   "SYLT",
				FrameTypeSyncTempoCodes:               "SYTC",
				FrameTypeTextAlbumName:                "TALB",
				FrameTypeTextBPM:                      "TBPM",
				FrameTypeTextCompilationItunes:        "TCMP",
				FrameTypeTextComposer:                 "TCOM",
				FrameTypeTextGenre:                    "TCON",
				FrameTypeTextCopyright:                "TCOP",
				FrameTypeTextDate:                     "TDAT",
				FrameTypeTextPlaylistDelay:            "TDLY",
				FrameTypeTextEncodedBy:                "TENC",
				FrameTypeTextLyricist:                 "TEXT",
				FrameTypeTextFileType:                 "TFLT",
				FrameTypeTextTime:                     "TIME",
				FrameTypeTextGroupDescription:         "TIT1",
				FrameTypeTextSongTitle:                "TIT2",
				FrameTypeTextSongSubtitle:             "TIT3",
				FrameTypeTextMusicalKey:               "TKEY",
				FrameTypeTextLanguage:                 "TLAN",
				FrameTypeTextLengthInMs:               "TLEN",
				FrameTypeTextMediaType:                "TMED",
				FrameTypeTextOriginalAlbum:            "TOAL",
				FrameTypeTextOriginalFileName:         "TOFN",
				FrameTypeTextOriginalLyricist:         "TOLY",
				FrameTypeTextOriginalPerformer:        "TOPE",
				FrameTypeTextOriginalReleaseTime:      "TORY",
				FrameTypeTextOwner:                    "TOWN",
				FrameTypeTextArtist:                   "TPE1",
				FrameTypeTextAlbumArtist:              "TPE2",
				FrameTypeTextConductor:                "TPE3",
				FrameTypeTextRemixer:                  "TPE4",
				FrameTypeTextPartOfSet:                "TPOS",
				FrameTypeTextPublisher:                "TPUB",
				FrameTypeTextTrackNumber:              "TRCK",
				FrameTypeTextRecordingDates:           "TRDA",
				FrameTypeTextRadioStation:             "TRSN",
				FrameTypeTextRadioStationOwner:        "TRSO",
				FrameTypeTextSize:                     "TSIZ",
				FrameTypeTextAlbumSortOrderItunes:     "TSO2",
				FrameTypeTextComposerSortOrderItunes:  "TSOC",
				FrameTypeTextISRC:                     "TSRC",
				FrameTypeTextEncodingSoftware:         "TSSE",
				FrameTypeTextRecordingTime:            "TYER",
				FrameTypeTextCustom:                   "TXXX",
				FrameTypeUniqueFileID:                 "UFID",
				FrameTypeTermsOfUse:                   "USER",
				FrameTypeLyricsUnsync:                 "USLT",
				FrameTypeURLCommercial:                "WCOM",
				FrameTypeURLCopyright:                 "WCOP",
				FrameTypeURLAudioFile:                 "WOAF",
				FrameTypeURLArtist:                    "WOAR",
				FrameTypeURLAudioSource:               "WOAS",
				FrameTypeURLRadioStation:              "WORS",
				FrameTypeURLPayment:                   "WPAY",
				FrameTypeURLPublisher:                 "WPUB",
				FrameTypeURLCustom:                    "WXXX",
				FrameTypeUnknown:                      "ZZZZ",
			}),
		}
	})

	return &codec23{vdata: v23Data}
}

func (c *codec23) Decode(t *Tag, r *reader) error {
	// Load the remaining six bytes of the tag header.
	if r.Load(6); r.err != nil {
		return r.err
	}

	// Decode the header.
	hdr := r.ConsumeBytes(10)
	if hdr[4] != 0 {
		return ErrInvalidTag
	}

	// Process tag header flags.
	flags := uint32(hdr[5])
	t.Flags = TagFlags(c.vdata.headerFlags.Decode(flags))

	// Process tag size.
	size, err := decodeSyncSafeUint32(hdr[6:10])
	if err != nil {
		return err
	}
	t.Size = int(size)

	// Load the rest of the tag into the reader's buffer.
	if r.Load(t.Size); r.err != nil {
		return r.err
	}

	// Remove unsync codes.
	if (t.Flags & TagFlagUnsync) != 0 {
		newBuf := removeUnsyncCodes(r.ConsumeAll())
		r.ReplaceBuffer(newBuf)
	}

	// Decode the extended header.
	if (t.Flags & TagFlagExtended) != 0 {
		exSize := decodeUint32(r.ConsumeBytes(4))

		// Decode the extended header flags.
		exFlags := r.ConsumeBytes(2)
		t.Flags = TagFlags(uint32(t.Flags) | c.vdata.headerExFlags.Decode(uint32(exFlags[0])<<8))

		// Consume the rest of the extended header data.
		exBytesConsumed := 6

		if (t.Flags & TagFlagHasCRC) != 0 {
			t.CRC = decodeUint32(r.ConsumeBytes(4))
			exBytesConsumed += 4
		}

		// Consume and ignore any remaining bytes in the extended header.
		if exBytesConsumed < int(exSize) {
			r.ConsumeBytes(int(exSize) - exBytesConsumed)
		}

		if r.err != nil {
			return r.err
		}
	}

	// Validate the CRC.
	if (t.Flags & TagFlagHasCRC) != 0 {
		crc := crc32.ChecksumIEEE(r.Bytes())
		if crc != t.CRC {
			return ErrFailedCRC
		}
	}

	// Decode the tag's frames until tag data is exhausted or padding is
	// encountered.
	for r.Len() > 0 {
		var f Frame
		err = c.decodeFrame(t, &f, r)

		if err == errPaddingEncountered {
			t.Padding = r.Len() + 4
			r.ConsumeAll()
			break
		}

		if err != nil {
			return err
		}

		t.Frames = append(t.Frames, f)
	}

	return nil
}

func (c *codec23) decodeFrame(t *Tag, f *Frame, r *reader) error {
	// Read the first four bytes of the frame header data to see if it's
	// padding.
	id := r.ConsumeBytes(4)
	if r.err != nil {
		return r.err
	}
	if id[0] == 0 && id[1] == 0 && id[2] == 0 && id[3] == 0 {
		return errPaddingEncountered
	}

	// Read the remaining 6 bytes of the header data into a buffer.
	hd := r.ConsumeBytes(6)
	if r.err != nil {
		return r.err
	}

	// Decode the frame's payload size.
	size := decodeUint32(hd[0:4])
	if size < 1 {
		return ErrInvalidFrameHeader
	}

	// Decode the frame flags.
	flags := c.vdata.frameFlags.Decode(uint32(hd[4])<<8 | uint32(hd[5]))

	// Start bulding the frame header.
	h := FrameHeader{
		FrameID: string(id),
		Size:    int(size),
		Flags:   FrameFlags(flags),
	}

	// Consume the rest of the frame into a new reader.
	r = r.ConsumeIntoNewReader(h.Size)

	// Decode extra header data.
	if h.Flags != 0 {
		if (h.Flags & FrameFlagCompressed) != 0 {
			h.DataLength = decodeUint32(r.ConsumeBytes(4))
		}

		if (h.Flags & FrameFlagEncrypted) != 0 {
			em := r.ConsumeByte()
			if em < 0x80 {
				return ErrInvalidEncryptMethod
			}
			h.EncryptMethod = em
		}

		if (h.Flags & FrameFlagHasGroupID) != 0 {
			gid := r.ConsumeByte()
			if gid < 0x80 {
				return ErrInvalidGroupID
			}
			h.GroupID = gid
		}

		if r.err != nil {
			return r.err
		}
	}

	// Use a reflector to scan the frame's fields.
	rf := newReflector(Version2_3, c.vdata)
	var err error
	*f, err = rf.ScanFrame(r, h.FrameID)
	if err != nil {
		return err
	}

	// Copy the header into the frame.
	rf.SetFrameHeader(*f, &h)
	return nil
}

func (c *codec23) Encode(t *Tag, w *writer) error {
	if (t.Flags & TagFlagHasCRC) != 0 {
		t.Flags |= TagFlagExtended
	}

	// Encode the header, leaving a placeholder for the size.
	flags := uint8(c.vdata.headerFlags.Encode(uint32(t.Flags)))
	hdr := []byte{'I', 'D', '3', 3, 0, flags, 0, 0, 0, 0}
	w.StoreBytes(hdr)
	sizeOffset := 6

	// Store the extended tag header.
	crcOffset := -1
	if (t.Flags & TagFlagExtended) != 0 {
		exFlags := uint16(c.vdata.headerExFlags.Encode(uint32(t.Flags)))

		// Store the first 6 bytes of the extended tag header, with a
		// placeholder for the extended header's size.
		exHdrOffset := w.Len()
		w.StoreBytes([]byte{0, 0, 0, 0, byte(exFlags >> 8), 0})

		if (t.Flags & TagFlagHasCRC) != 0 {
			crcOffset = w.Len()
			w.StoreBytes([]byte{0, 0, 0, 0})
		}

		// Update the extended header size.
		exSize := w.Len() - exHdrOffset
		encodeUint32(w.SliceBuffer(exHdrOffset, 4), uint32(exSize))
	}

	// Encode the frames.
	framesOffset := w.Len()
	for _, f := range t.Frames {
		if err := c.encodeFrame(t, f, w); err != nil {
			return err
		}
	}

	// Add padding.
	if t.Padding > 0 {
		if t.Padding < 4 {
			t.Padding = 4 // must be at least 4 bytes.
		}
		w.StoreBytes(make([]byte, t.Padding))
	}

	// Calculate a CRC covering only the frames and padding, and store it into
	// the extended header.
	if crcOffset > -1 {
		framesBuf := w.SliceBuffer(framesOffset, w.Len()-framesOffset)
		t.CRC = uint32(crc32.ChecksumIEEE(framesBuf))
		crcBuf := w.SliceBuffer(crcOffset, 4)
		encodeUint32(crcBuf, t.CRC)
	}

	// Unsynchronize.
	if (t.Flags & TagFlagUnsync) != 0 {
		b := addUnsyncCodes(w.ConsumeBytesFromOffset(10))
		w.StoreBytes(b)
	}

	// Update the tag header's size.
	t.Size = w.Len() - len(hdr)
	sizeBuf := w.SliceBuffer(sizeOffset, 4)
	encodeSyncSafeUint32(sizeBuf, uint32(t.Size))

	// Save writer's buffer to the output stream.
	_, err := w.Save()
	return err
}

func (c *codec23) encodeFrame(t *Tag, f Frame, w *writer) error {
	// Store a placeholder for the frame ID.
	idOffset := w.Len()
	w.StoreBytes([]byte{0, 0, 0, 0})

	// Store a placeholder for the frame size.
	sizeOffset := w.Len()
	w.StoreBytes([]byte{0, 0, 0, 0})

	// Retrieve the frame's header.
	h := HeaderOf(f)

	// Encode the frame header flags.
	flags := c.vdata.frameFlags.Encode(uint32(h.Flags))
	w.StoreByte(byte(flags >> 8))
	w.StoreByte(byte(flags))

	// Encode additional header data indicated by header flags.
	startOffset := w.Len()
	dataLengthOffset := -1
	if h.Flags != 0 {
		if (h.Flags & FrameFlagCompressed) != 0 {
			dataLengthOffset = w.Len()
			w.StoreBytes([]byte{0, 0, 0, 0})
		}

		if (h.Flags & FrameFlagEncrypted) != 0 {
			if h.EncryptMethod < 0x80 {
				w.err = ErrInvalidEncryptMethod
			}
			w.StoreByte(h.EncryptMethod)
		}

		if (h.Flags & FrameFlagHasGroupID) != 0 {
			if h.GroupID < 0x80 {
				w.err = ErrInvalidGroupID
			}
			w.StoreByte(h.GroupID)
		}

		if w.err != nil {
			return w.err
		}
	}

	payloadOffset := w.Len()

	// Use a reflector to output the frame's fields.
	rf := newReflector(Version2_3, c.vdata)
	frameID, err := rf.OutputFrame(w, f)
	if err != nil {
		return err
	}

	// Update data length.
	if dataLengthOffset > -1 {
		dl := w.Len() - payloadOffset
		encodeUint32(w.SliceBuffer(dataLengthOffset, 4), uint32(dl))
	}

	// Update the header frame ID.
	h.FrameID = frameID
	copy(w.SliceBuffer(idOffset, 4), []byte(h.FrameID))

	// Update the header frame size.
	h.Size = w.Len() - startOffset
	encodeSyncSafeUint32(w.SliceBuffer(sizeOffset, 4), uint32(h.Size))

	return w.err
}
