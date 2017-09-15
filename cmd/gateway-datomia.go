package cmd

import (
	"io"

	"github.com/istreamdata/istreamdatago/datweb/dwebclient"
)

type datomiaObjects struct {
	cli *dwebclient.Client
}

func newDatomiaGateway(addr, vault string) (ObjectLayer, error) {
	return &datomiaObjects{
		cli: dwebclient.NewClient(addr, vault),
	}, nil
}

func (d *datomiaObjects) Shutdown() error {
	return nil
}

func (d *datomiaObjects) StorageInfo() StorageInfo {
	return StorageInfo{}
}

func (d *datomiaObjects) MakeBucketWithLocation(bucket string, location string) error {
	if !IsValidBucketName(bucket) {
		return traceError(BucketNameInvalid{Bucket: bucket})
	}

	panic("not implemented")
}

func (d *datomiaObjects) GetBucketInfo(bucket string) (bucketInfo BucketInfo, err error) {
	panic("not implemented")
}

func (d *datomiaObjects) ListBuckets() (buckets []BucketInfo, err error) {
	panic("not implemented")
}

func (d *datomiaObjects) DeleteBucket(bucket string) error {
	panic("not implemented")
}

func (d *datomiaObjects) ListObjects(bucket string, prefix string, marker string, delimiter string, maxKeys int) (result ListObjectsInfo, err error) {
	panic("not implemented")
}

func (d *datomiaObjects) GetObject(bucket string, object string, startOffset int64, length int64, writer io.Writer) (err error) {
	panic("not implemented")
}

func (d *datomiaObjects) GetObjectInfo(bucket string, object string) (objInfo ObjectInfo, err error) {
	panic("not implemented")
}

func (d *datomiaObjects) PutObject(bucket string, object string, size int64, data io.Reader, metadata map[string]string, sha256sum string) (objInfo ObjectInfo, err error) {
	panic("not implemented")
}

func (d *datomiaObjects) CopyObject(srcBucket string, srcObject string, destBucket string, destObject string, metadata map[string]string) (objInfo ObjectInfo, err error) {
	panic("not implemented")
}

func (d *datomiaObjects) DeleteObject(bucket string, object string) error {
	panic("not implemented")
}

func (d *datomiaObjects) ListMultipartUploads(bucket string, prefix string, keyMarker string, uploadIDMarker string, delimiter string, maxUploads int) (result ListMultipartsInfo, err error) {
	panic("not implemented")
}

func (d *datomiaObjects) NewMultipartUpload(bucket string, object string, metadata map[string]string) (uploadID string, err error) {
	panic("not implemented")
}

func (d *datomiaObjects) CopyObjectPart(srcBucket string, srcObject string, destBucket string, destObject string, uploadID string, partID int, startOffset int64, length int64) (info PartInfo, err error) {
	return PartInfo{}, traceError(NotSupported{})
}

func (d *datomiaObjects) PutObjectPart(bucket string, object string, uploadID string, partID int, size int64, data io.Reader, md5Hex string, sha256sum string) (info PartInfo, err error) {
	panic("not implemented")
}

func (d *datomiaObjects) ListObjectParts(bucket string, object string, uploadID string, partNumberMarker int, maxParts int) (result ListPartsInfo, err error) {
	panic("not implemented")
}

func (d *datomiaObjects) AbortMultipartUpload(bucket string, object string, uploadID string) error {
	panic("not implemented")
}

func (d *datomiaObjects) CompleteMultipartUpload(bucket string, object string, uploadID string, uploadedParts []completePart) (objInfo ObjectInfo, err error) {
	panic("not implemented")
}

func (d *datomiaObjects) HealBucket(bucket string) error {
	return traceError(NotImplemented{})
}

func (d *datomiaObjects) ListBucketsHeal() (buckets []BucketInfo, err error) {
	return nil, traceError(NotImplemented{})
}

func (d *datomiaObjects) HealObject(bucket string, object string) (int, int, error) {
	return 0, 0, traceError(NotImplemented{})
}

func (d *datomiaObjects) ListObjectsHeal(bucket string, prefix string, marker string, delimiter string, maxKeys int) (ListObjectsInfo, error) {
	return ListObjectsInfo{}, traceError(NotImplemented{})
}

func (d *datomiaObjects) ListUploadsHeal(bucket string, prefix string, marker string, uploadIDMarker string, delimiter string, maxUploads int) (ListMultipartsInfo, error) {
	return ListMultipartsInfo{}, traceError(NotImplemented{})
}
