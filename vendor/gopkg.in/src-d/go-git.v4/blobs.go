package git

import (
	"io"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

// Blob is used to store file data - it is generally a file.
type Blob struct {
	Hash plumbing.Hash
	Size int64

	obj plumbing.Object
}

// ID returns the object ID of the blob. The returned value will always match
// the current value of Blob.Hash.
//
// ID is present to fulfill the Object interface.
func (b *Blob) ID() plumbing.Hash {
	return b.Hash
}

// Type returns the type of object. It always returns plumbing.BlobObject.
//
// Type is present to fulfill the Object interface.
func (b *Blob) Type() plumbing.ObjectType {
	return plumbing.BlobObject
}

// Decode transforms a plumbing.Object into a Blob struct.
func (b *Blob) Decode(o plumbing.Object) error {
	if o.Type() != plumbing.BlobObject {
		return ErrUnsupportedObject
	}

	b.Hash = o.Hash()
	b.Size = o.Size()
	b.obj = o

	return nil
}

// Encode transforms a Blob into a plumbing.Object.
func (b *Blob) Encode(o plumbing.Object) error {
	w, err := o.Writer()
	if err != nil {
		return err
	}
	defer checkClose(w, &err)
	r, err := b.Reader()
	if err != nil {
		return err
	}
	defer checkClose(r, &err)
	_, err = io.Copy(w, r)
	o.SetType(plumbing.BlobObject)
	return err
}

// Reader returns a reader allow the access to the content of the blob
func (b *Blob) Reader() (io.ReadCloser, error) {
	return b.obj.Reader()
}

// BlobIter provides an iterator for a set of blobs.
type BlobIter struct {
	storer.ObjectIter
	r *Repository
}

// NewBlobIter returns a CommitIter for the given repository and underlying
// object iterator.
//
// The returned BlobIter will automatically skip over non-blob objects.
func NewBlobIter(r *Repository, iter storer.ObjectIter) *BlobIter {
	return &BlobIter{iter, r}
}

// Next moves the iterator to the next blob and returns a pointer to it. If it
// has reached the end of the set it will return io.EOF.
func (iter *BlobIter) Next() (*Blob, error) {
	for {
		obj, err := iter.ObjectIter.Next()
		if err != nil {
			return nil, err
		}

		if obj.Type() != plumbing.BlobObject {
			continue
		}

		blob := &Blob{}
		return blob, blob.Decode(obj)
	}
}

// ForEach call the cb function for each blob contained on this iter until
// an error happens or the end of the iter is reached. If ErrStop is sent
// the iteration is stop but no error is returned. The iterator is closed.
func (iter *BlobIter) ForEach(cb func(*Blob) error) error {
	return iter.ObjectIter.ForEach(func(obj plumbing.Object) error {
		if obj.Type() != plumbing.BlobObject {
			return nil
		}

		blob := &Blob{}
		if err := blob.Decode(obj); err != nil {
			return err
		}

		return cb(blob)
	})
}
