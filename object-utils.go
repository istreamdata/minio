/*
 * Minio Cloud Storage, (C) 2015, 2016 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/minio/minio/pkg/safe"
)

const (
	// Minio meta bucket.
	minioMetaBucket = ".minio"
)

// validBucket regexp.
var validBucket = regexp.MustCompile(`^[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]$`)

// IsValidBucketName verifies a bucket name in accordance with Amazon's
// requirements. It must be 3-63 characters long, can contain dashes
// and periods, but must begin and end with a lowercase letter or a number.
// See: http://docs.aws.amazon.com/AmazonS3/latest/dev/BucketRestrictions.html
func IsValidBucketName(bucket string) bool {
	if len(bucket) < 3 || len(bucket) > 63 {
		return false
	}
	if bucket[0] == '.' || bucket[len(bucket)-1] == '.' {
		return false
	}
	return validBucket.MatchString(bucket)
}

// IsValidObjectName verifies an object name in accordance with Amazon's
// requirements. It cannot exceed 1024 characters and must be a valid UTF8
// string.
//
// See:
// http://docs.aws.amazon.com/AmazonS3/latest/dev/UsingMetadata.html
//
// You should avoid the following characters in a key name because of
// significant special handling for consistency across all
// applications.
//
// Rejects strings with following characters.
//
// - Backslash ("\")
// - Caret ("^")
// - Grave accent / back tick ("`")
// - Vertical bar / pipe ("|")
func IsValidObjectName(object string) bool {
	if len(object) > 1024 || len(object) == 0 {
		return false
	}
	if !utf8.ValidString(object) {
		return false
	}
	// Reject unsupported characters in object name.
	return !strings.ContainsAny(object, "`^*|\\\"")
}

// IsValidObjectPrefix verifies whether the prefix is a valid object name.
// Its valid to have a empty prefix.
func IsValidObjectPrefix(object string) bool {
	// Prefix can be empty or "/".
	if object == "" || object == "/" {
		return true
	}
	// Verify if prefix is a valid object name.
	return IsValidObjectName(object)

}

// Slash separator.
const slashSeparator = "/"

// retainSlash - retains slash from a path.
func retainSlash(s string) string {
	return strings.TrimSuffix(s, slashSeparator) + slashSeparator
}

// pathJoin - path join.
func pathJoin(s1 string, s2 string) string {
	return retainSlash(s1) + s2
}

// Create an s3 compatible MD5sum for complete multipart transaction.
func makeS3MD5(md5Strs ...string) (string, error) {
	var finalMD5Bytes []byte
	for _, md5Str := range md5Strs {
		md5Bytes, err := hex.DecodeString(md5Str)
		if err != nil {
			return "", err
		}
		finalMD5Bytes = append(finalMD5Bytes, md5Bytes...)
	}
	md5Hasher := md5.New()
	md5Hasher.Write(finalMD5Bytes)
	s3MD5 := fmt.Sprintf("%s-%d", hex.EncodeToString(md5Hasher.Sum(nil)), len(md5Strs))
	return s3MD5, nil
}

// byBucketName is a collection satisfying sort.Interface.
type byBucketName []BucketInfo

func (d byBucketName) Len() int           { return len(d) }
func (d byBucketName) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d byBucketName) Less(i, j int) bool { return d[i].Name < d[j].Name }

// safeCloseAndRemove - safely closes and removes underlying temporary
// file writer if possible.
func safeCloseAndRemove(writer io.WriteCloser) error {
	// If writer is a safe file, Attempt to close and remove.
	safeWriter, ok := writer.(*safe.File)
	if ok {
		return safeWriter.CloseAndRemove()
	}
	pipeWriter, ok := writer.(*io.PipeWriter)
	if ok {
		return pipeWriter.CloseWithError(errors.New("Close and error out."))
	}
	return nil
}
