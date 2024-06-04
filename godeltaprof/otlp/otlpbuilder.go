package otlp

import (
	"github.com/grafana/pyroscope-go/godeltaprof/internal/pprof"
	otlpprofile "go.opentelemetry.io/proto/otlp/profiles/v1experimental"

	"runtime"
	"time"
)

// This is a copy of profileBuilder from godeltaprof/internal/pprof/proto.go
// which in turn is a copy of profileBuilder from runtime/pprof/internal/pprof/proto.go
// otlpProtoBuilder is a builder for the OTLP profile proto format.
type otlpProtoBuilder struct {
	stc  pprof.ProfileConfig
	mem  []pprof.MemMap
	deck pprof.PCDeck

	locIndex  []int64
	strings   []string
	stringMap map[string]int
	locs      map[uintptr]locInfo // list of locInfo starting with the given PC.
	funcs     map[string]uint64   // Package path-qualified function name to Function.ID

	opt *pprof.ProfileBuilderOptions

	res     *otlpprofile.Profile
	tmpLocs []uint64
	start   time.Time
}

func newOTLPProtoBuilder(stc pprof.ProfileConfig, opt *pprof.ProfileBuilderOptions) *otlpProtoBuilder {
	b := &otlpProtoBuilder{
		start: time.Now(),
		stc:   stc,
		opt:   opt,
		mem:   opt.Mapping(),
		res:   &otlpprofile.Profile{},

		strings:   []string{""},
		stringMap: map[string]int{"": 0},
		locs:      map[uintptr]locInfo{},
		funcs:     map[string]uint64{},
	}
	b.res.PeriodType = &otlpprofile.ValueType{
		Type: b.stringIndex(stc.PeriodType.Typ),
		Unit: b.stringIndex(stc.PeriodType.Unit),
	}
	b.res.Period = stc.Period
	for _, sampleType := range stc.SampleType {
		b.res.SampleType = append(b.res.SampleType, &otlpprofile.ValueType{
			Type: b.stringIndex(sampleType.Typ),
			Unit: b.stringIndex(sampleType.Unit),
		})
	}
	if stc.DefaultSampleType != "" {
		b.res.DefaultSampleType = b.stringIndex(stc.DefaultSampleType)
	}

	return b
}

func (b *otlpProtoBuilder) LocsForStack(stk []uintptr) (newLocs []uint64) {
	return b.appendLocsForStack(stk)
}

func (b *otlpProtoBuilder) Sample(values []int64, locs []uint64, _ int64) {
	start := len(b.locIndex)
	sz := len(locs)
	for _, loc := range locs { //todo
		b.locIndex = append(b.locIndex, int64(loc))
	}
	vs := make([]int64, len(values))
	copy(vs, values)
	b.res.Sample = append(b.res.Sample, &otlpprofile.Sample{
		LocationIndex:       nil,
		LocationsStartIndex: uint64(start),
		LocationsLength:     uint64(sz),
		StacktraceIdIndex:   0,
		Value:               vs,
		Label:               nil,
		Attributes:          nil,
		Link:                0,
		TimestampsUnixNano:  nil,
	})

}

func (b *otlpProtoBuilder) Build() error {
	b.res.StringTable = b.strings
	b.res.LocationIndices = b.locIndex
	b.res.TimeNanos = b.start.UnixNano()
	b.res.DurationNanos = time.Since(b.start).Nanoseconds()
	b.res.Mapping = make([]*otlpprofile.Mapping, 0, len(b.mem))
	for i, m := range b.mem {
		hasFunctions := m.Funcs == pprof.LookupTried // LookupTried but not LookupFailed
		b.res.Mapping = append(b.res.Mapping, &otlpprofile.Mapping{
			Id:           uint64(i + 1),
			MemoryStart:  uint64(m.Start),
			MemoryLimit:  uint64(m.End),
			FileOffset:   m.Offset,
			Filename:     b.stringIndex(m.File),
			BuildId:      b.stringIndex(m.BuildID),
			BuildIdKind:  otlpprofile.BuildIdKind_BUILD_ID_LINKER,
			HasFunctions: hasFunctions,
		})
	}
	return nil
}

func (b *otlpProtoBuilder) Proto() *otlpprofile.Profile {
	return b.res
}

// appendLocsForStack appends the location IDs for the given stack trace to the given
// location ID slice, locs. The addresses in the stack are return PCS or 1 + the PC of
// an inline marker as the runtime traceback function returns.
//
// It may return an empty slice even if locs is non-empty, for example if locs consists
// solely of runtime.goexit. We still count these empty stacks in profiles in order to
// get the right cumulative sample count.
//
// It may emit to b.pb, so there must be no message encoding in progress.
func (b *otlpProtoBuilder) appendLocsForStack(stk []uintptr) (newLocs []uint64) {
	b.deck.Reset()
	locs := b.tmpLocs[:0]

	// The last frame might be truncated. Recover lost inline Frames.
	stk = pprof.Runtime_expandFinalInlineFrame(stk)

	for len(stk) > 0 {
		addr := stk[0]
		if l, ok := b.locs[addr]; ok {
			// When generating code for an inlined function, the compiler adds
			// NOP instructions to the outermost function as a placeholder for
			// each layer of inlining. When the runtime generates tracebacks for
			// stacks that include inlined functions, it uses the addresses of
			// those NOPs as "fake" PCS on the stack as if they were regular
			// function call sites. But if a profiling signal arrives while the
			// CPU is executing one of those NOPs, its PC will show up as a leaf
			// in the profile with its own Location entry. So, always check
			// whether addr is a "fake" PC in the context of the current call
			// stack by trying to add it to the inlining deck before assuming
			// that the deck is complete.
			if len(b.deck.PCS) > 0 {
				if added := b.deck.TryAdd(addr, l.firstPCFrames, l.firstPCSymbolizeResult); added {
					stk = stk[1:]
					continue
				}
			}

			// first record the location if there is any pending accumulated info.
			if id, ok := b.emitLocation(); ok {
				locs = append(locs, id)
			}

			// then, record the cached location.
			locs = append(locs, l.id)

			// Skip the matching PCS.
			//
			// Even if stk was truncated due to the stack depth
			// limit, expandFinalInlineFrame above has already
			// fixed the truncation, ensuring it is long enough.
			stk = stk[len(l.pcs):]
			continue
		}

		frames, symbolizeResult := pprof.AllFrames(addr)
		if len(frames) == 0 { // runtime.goexit.
			if id, ok := b.emitLocation(); ok {
				locs = append(locs, id)
			}
			stk = stk[1:]
			continue
		}

		if added := b.deck.TryAdd(addr, frames, symbolizeResult); added {
			stk = stk[1:]
			continue
		}
		// add failed because this addr is not inlined with the
		// existing PCS in the deck. Flush the deck and retry handling
		// this pc.
		if id, ok := b.emitLocation(); ok {
			locs = append(locs, id)
		}

		// check cache again - previous emitLocation added a new entry
		if l, ok := b.locs[addr]; ok {
			locs = append(locs, l.id)
			stk = stk[len(l.pcs):] // skip the matching PCS.
		} else {
			b.deck.TryAdd(addr, frames, symbolizeResult) // must succeed.
			stk = stk[1:]
		}
	}
	if id, ok := b.emitLocation(); ok { // emit remaining location.
		locs = append(locs, id)
	}
	b.tmpLocs = locs
	return locs
}

// emitLocation emits the new location and function information recorded in the deck
// and returns the location ID encoded in the profile protobuf.
// It emits to b.pb, so there must be no message encoding in progress.
// It resets the deck.
func (b *otlpProtoBuilder) emitLocation() (uint64, bool) {
	if len(b.deck.PCS) == 0 {
		return 0, false
	}
	defer b.deck.Reset()

	addr := b.deck.PCS[0]
	firstFrame := b.deck.Frames[0]

	// We can't write out functions while in the middle of the
	// Location message, so record new functions we encounter and
	// write them out after the Location.
	type newFunc struct {
		id         uint64
		name, file string
		startLine  int64
	}
	newFuncs := make([]newFunc, 0, 8)

	id := uint64(len(b.locs))
	b.locs[addr] = locInfo{
		id:                     id,
		pcs:                    append([]uintptr{}, b.deck.PCS...),
		firstPCSymbolizeResult: b.deck.FirstPCSymbolizeResult,
		firstPCFrames:          append([]runtime.Frame{}, b.deck.Frames[:b.deck.FirstPCFrames]...),
	}

	loc := &otlpprofile.Location{
		Id:      id,
		Address: uint64(firstFrame.PC),
		Line:    make([]*otlpprofile.Line, 0, len(b.deck.Frames)),
	}
	for _, frame := range b.deck.Frames {
		// Write out each line in frame expansion.
		funcName := pprof.Runtime_FrameSymbolName(&frame)
		funcID, ok := b.funcs[funcName]
		if !ok {
			funcID = uint64(len(b.funcs))
			b.funcs[frame.Function] = funcID
			var name string
			if b.opt.GenericsFrames {
				name = funcName
			} else {
				name = frame.Function
			}
			newFuncs = append(newFuncs, newFunc{
				id:        funcID,
				name:      name,
				file:      frame.File,
				startLine: int64(pprof.Runtime_FrameStartLine(&frame)),
			})
		}
		loc.Line = append(loc.Line, &otlpprofile.Line{
			FunctionIndex: funcID,
			Line:          int64(frame.Line),
			Column:        0,
		})
	}
	for i := range b.mem {
		if b.mem[i].Start <= addr && addr < b.mem[i].End || b.mem[i].Fake {
			loc.MappingIndex = uint64(i)

			m := b.mem[i]
			m.Funcs |= b.deck.SymbolizeResult
			b.mem[i] = m
			break
		}
	}

	// Write out functions we found during frame expansion.
	for _, fn := range newFuncs {
		b.res.Function = append(b.res.Function, &otlpprofile.Function{
			Id:         fn.id,
			Name:       b.stringIndex(fn.name),
			SystemName: b.stringIndex(fn.name),
			Filename:   b.stringIndex(fn.file),
			StartLine:  fn.startLine,
		})
	}
	b.res.Location = append(b.res.Location, loc)
	return id, true
}

func (b *otlpProtoBuilder) stringIndex(s string) int64 {
	id, ok := b.stringMap[s]
	if !ok {
		id = len(b.strings)
		b.strings = append(b.strings, s)
		b.stringMap[s] = id
	}
	return int64(id)
}

type locInfo struct {
	// location id assigned by the PPROFProfileBuilder
	id uint64

	// sequence of PCS, including the Fake PCS returned by the traceback
	// to represent inlined functions
	// https://github.com/golang/go/blob/d6f2f833c93a41ec1c68e49804b8387a06b131c5/src/runtime/traceback.go#L347-L368
	pcs []uintptr

	// firstPCFrames and firstPCSymbolizeResult hold the results of the
	// AllFrames call for the first (leaf-most) PC this locInfo represents
	firstPCFrames          []runtime.Frame
	firstPCSymbolizeResult pprof.SymbolizeFlag
}
