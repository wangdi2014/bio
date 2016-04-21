package seq

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/shenwei356/util/byteutil"
)

// Seq struct has two attributes, alphabet, seq,
type Seq struct {
	Alphabet  *Alphabet
	Seq       []byte
	Qual      []byte
	QualValue []int
}

// ValidateSeq decides whether check sequence or not
var ValidateSeq = true

// NewSeq is constructor for type *Seq*
func NewSeq(t *Alphabet, s []byte) (*Seq, error) {
	if ValidateSeq {
		// check sequene first
		if err := t.IsValid(s); err != nil {
			return nil, err
		}
	}

	seq := &Seq{Alphabet: t, Seq: s}
	return seq, nil
}

// NewSeqWithQual is used to store fastq sequence
func NewSeqWithQual(t *Alphabet, s []byte, q []byte) (*Seq, error) {
	if len(s) != len(q) {
		return nil, fmt.Errorf("unmatched length of sequence (%d) and quality (%d)", len(s), len(q))
	}
	seq, err := NewSeq(t, s)
	if err != nil {
		return nil, err
	}
	seq.Qual = q
	return seq, nil
}

// NewSeqWithoutValidate create Seq without check the sequences
func NewSeqWithoutValidate(t *Alphabet, s []byte) (*Seq, error) {
	seq := &Seq{Alphabet: t, Seq: s}
	return seq, nil
}

// NewSeqWithQualWithoutValidate create Seq with quality without check the sequences
func NewSeqWithQualWithoutValidate(t *Alphabet, s []byte, q []byte) (*Seq, error) {
	if len(s) != len(q) {
		return nil, fmt.Errorf("unmatched length of sequence (%d) and quality (%d)", len(s), len(q))
	}
	seq := &Seq{Alphabet: t, Seq: s, Qual: q}
	return seq, nil
}

// Length returns the lenght of sequence
func (seq *Seq) Length() int {
	return len(seq.Seq)
}

// SubSeq returns a sub seq. start and end is 1-based.
// end could be below than 0, e.g. SubSeq(1, -2) return
// seq without the last base.
func (seq *Seq) SubSeq(start int, end int) *Seq {
	if start < 1 {
		start = 1
	}
	if end > len(seq.Seq) {
		end = len(seq.Seq)
	}
	if end < 1 {
		if end == 0 {
			end = -1
		}
		end = len(seq.Seq) + end - 1
	}
	newseq, _ := NewSeqWithoutValidate(seq.Alphabet, seq.Seq[start-1:end])
	if len(seq.Qual) > 0 {
		newseq.Qual = seq.Qual[start-1 : end]
	}
	if len(seq.QualValue) > 0 {
		newseq.QualValue = seq.QualValue[start-1 : end]
	}
	return newseq
}

// RemoveGaps remove gaps
func (seq *Seq) RemoveGaps(letters string) *Seq {
	if len(letters) == 0 {
		newseq, _ := NewSeqWithQualWithoutValidate(seq.Alphabet, seq.Seq, seq.Qual)
		return newseq
	}

	// do not use map
	querySlice := make([]byte, 256)
	for i := 0; i < len(letters); i++ {
		querySlice[int(letters[i])] = letters[i]
	}

	s := []byte{}
	q := []byte{}
	var b, g byte
	for i := 0; i < len(seq.Seq); i++ {
		b = seq.Seq[i]

		g = querySlice[int(b)]
		if g == 0 {
			s = append(s, b)
			if len(seq.Qual) > 0 {
				q = append(q, seq.Qual[i])
			}
		}
	}
	newseq, _ := NewSeqWithQualWithoutValidate(seq.Alphabet, s, q)
	return newseq
}

// RevCom returns reverse complement sequence
func (seq *Seq) RevCom() *Seq {
	return seq.Complement().Reverse()
}

// Reverse a sequence
func (seq *Seq) Reverse() *Seq {
	if len(seq.Qual) > 0 {
		s := byteutil.ReverseByteSlice(seq.Seq)
		newseq, _ := NewSeqWithQualWithoutValidate(seq.Alphabet, s, byteutil.ReverseByteSlice(seq.Qual))
		return newseq
	}
	s := byteutil.ReverseByteSlice(seq.Seq)
	newseq, _ := NewSeqWithoutValidate(seq.Alphabet, s)
	return newseq
}

// Complement returns complement sequence. Note that is will lose quality information
func (seq *Seq) Complement() *Seq {
	if seq.Alphabet == Unlimit {
		newseq, _ := NewSeqWithoutValidate(seq.Alphabet, []byte(""))
		return newseq
	}

	s := make([]byte, len(seq.Seq))
	var p byte
	for i := 0; i < len(seq.Seq); i++ {
		p, _ = seq.Alphabet.PairLetter(seq.Seq[i])
		s[i] = p
	}

	newseq, _ := NewSeqWithoutValidate(seq.Alphabet, s)
	return newseq
}

// FormatSeq wrap seq
func (seq *Seq) FormatSeq(width int) []byte {
	return byteutil.WrapByteSlice(seq.Seq, width)
}

/*BaseContent returns base content for given bases. For example:

  seq.BaseContent("gc")

*/
func (seq *Seq) BaseContent(list string) float64 {
	if len(seq.Seq) == 0 {
		return float64(0)
	}

	sum := 0
	for _, b := range []byte(list) {
		up := bytes.ToUpper([]byte{b})
		lo := bytes.ToLower([]byte{b})
		if string(up) == string(lo) {
			sum += bytes.Count(seq.Seq, up)
		} else {
			sum += bytes.Count(seq.Seq, up) + bytes.Count(seq.Seq, lo)
		}
	}

	return float64(sum) / float64(len(seq.Seq))
}

// GC returns the GC content
func (seq *Seq) GC() float64 {
	return seq.BaseContent("gc")
}

// DegenerateBaseMapNucl mappings nucleic acid degenerate base to
// regular expression
var DegenerateBaseMapNucl = map[byte]string{
	'A': "A",
	'T': "T",
	'U': "U",
	'C': "C",
	'G': "G",
	'R': "[AG]",
	'Y': "[CT]",
	'M': "[AC]",
	'K': "[GT]",
	'S': "[CG]",
	'W': "[AT]",
	'H': "[ACT]",
	'B': "[CGT]",
	'V': "[ACG]",
	'D': "[AGT]",
	'N': "[ACGT]",
	'a': "a",
	't': "t",
	'u': "u",
	'c': "c",
	'g': "g",
	'r': "[ag]",
	'y': "[ct]",
	'm': "[ac]",
	'k': "[gt]",
	's': "[cg]",
	'w': "[at]",
	'h': "[act]",
	'b': "[cgt]",
	'v': "[acg]",
	'd': "[agt]",
	'n': "[acgt]",
}

// DegenerateBaseMapProt mappings protein degenerate base to
// regular expression
var DegenerateBaseMapProt = map[byte]string{
	'A': "A",
	'B': "[DN]",
	'C': "C",
	'D': "D",
	'E': "E",
	'F': "F",
	'G': "G",
	'H': "H",
	'I': "I",
	'J': "[IL]",
	'K': "K",
	'L': "L",
	'M': "M",
	'N': "N",
	'P': "P",
	'Q': "Q",
	'R': "R",
	'S': "S",
	'T': "T",
	'V': "V",
	'W': "W",
	'Y': "Y",
	'Z': "[QE]",
	'a': "a",
	'b': "[dn]",
	'c': "c",
	'd': "d",
	'e': "e",
	'f': "f",
	'g': "g",
	'h': "h",
	'i': "i",
	'j': "[il]",
	'k': "k",
	'l': "l",
	'm': "m",
	'n': "n",
	'p': "p",
	'q': "q",
	'r': "r",
	's': "s",
	't': "t",
	'v': "v",
	'w': "w",
	'y': "y",
	'z': "[qe]",
}

// Degenerate2Regexp transform seqs containing degenrate base to regular expression
func (seq *Seq) Degenerate2Regexp() string {
	var m map[byte]string
	if seq.Alphabet == Protein {
		m = DegenerateBaseMapProt
	} else {
		m = DegenerateBaseMapNucl
	}

	s := make([]string, len(seq.Seq))
	for i, base := range seq.Seq {
		if _, ok := m[base]; ok {
			s[i] = m[base]
		} else {
			s[i] = string(base)
		}
	}
	return strings.Join(s, "")
}
