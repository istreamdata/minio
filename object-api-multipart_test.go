/*
 * Minio Cloud Storage, (C) 2016 Minio, Inc.
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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

// Tests validate creation of new multipart upload instance.
func TestObjectNewMultipartUpload(t *testing.T) {
	directory, err := ioutil.TempDir("", "minio-multipart-1-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(directory)

	// Initialize fs object layer.
	obj, err := newFSObjects(directory)
	if err != nil {
		t.Fatal(err)
	}

	bucket := "minio-bucket"
	object := "minio-object"

	errMsg := "Bucket not found: minio-bucket"
	// opearation expected to fail since the bucket on which NewMultipartUpload is being initiated doesn't exist.
	uploadID, err := obj.NewMultipartUpload(bucket, object)
	if err == nil {
		t.Fatalf("Expected to fail since the NewMultipartUpload is intialized on a non-existant bucket.")
	}
	if errMsg != err.Error() {
		t.Errorf("Expected to fail with Error \"%s\", but instead found \"%s\".", errMsg, err.Error())
	}

	// Create bucket before intiating NewMultipartUpload.
	err = obj.MakeBucket(bucket)
	if err != nil {
		// failed to create newbucket, abort.
		t.Fatal(err)
	}

	uploadID, err = obj.NewMultipartUpload(bucket, object)
	if err != nil {
		t.Fatal(err)
	}

	err = obj.AbortMultipartUpload(bucket, object, uploadID)
	if err != nil {
		switch err.(type) {
		case InvalidUploadID:
			t.Fatalf("New Multipart upload failed to create uuid file.")
		default:
			t.Fatalf(err.Error())
		}
	}
}

// Tests validates the validator for existence of uploadID.
func TestObjectAPIIsUploadIDExists(t *testing.T) {
	directory, err := ioutil.TempDir("", "minio-multipart-2-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(directory)

	// Initialize fs object layer.
	obj, err := newFSObjects(directory)
	if err != nil {
		t.Fatal(err)
	}

	bucket := "minio-bucket"
	object := "minio-object"

	// Create bucket before intiating NewMultipartUpload.
	err = obj.MakeBucket(bucket)
	if err != nil {
		// Failed to create newbucket, abort.
		t.Fatal(err)
	}

	_, err = obj.NewMultipartUpload(bucket, object)
	if err != nil {
		t.Fatal(err)
	}

	err = obj.AbortMultipartUpload(bucket, object, "abc")
	switch err.(type) {
	case InvalidUploadID:
	default:
		t.Fatal("Expected uploadIDPath to exist.")
	}
}

// Tests validate correctness of PutObjectPart.
func TestObjectAPIPutObjectPart(t *testing.T) {
	// Generating cases for which the PutObjectPart fails.
	directory, err := ioutil.TempDir("", "minio-multipart-3-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(directory)

	// Initializing fs object layer.
	obj, err := newFSObjects(directory)
	if err != nil {
		t.Fatal(err)
	}

	bucket := "minio-bucket"
	object := "minio-object"

	// Create bucket before intiating NewMultipartUpload.
	err = obj.MakeBucket(bucket)
	if err != nil {
		// Failed to create newbucket, abort.
		t.Fatal(err)
	}
	// Initiate Multipart Upload on the above created bucket.
	uploadID, err := obj.NewMultipartUpload(bucket, object)
	if err != nil {
		// Failed to create NewMultipartUpload, abort.
		t.Fatal(err)
	}
	// Creating a dummy bucket for tests.
	err = obj.MakeBucket("unused-bucket")
	if err != nil {
		// Failed to create newbucket, abort.
		t.Fatal(err)
	}

	failCases := []struct {
		bucketName      string
		objName         string
		uploadID        string
		PartID          int
		inputReaderData string
		inputMd5        string
		intputDataSize  int64
		// flag indicating whether the test should pass.
		shouldPass bool
		// expected error output.
		expectedMd5   string
		expectedError error
	}{
		// Test case  1-4.
		// Cases with invalid bucket name.
		{".test", "obj", "", 1, "", "", 0, false, "", fmt.Errorf("%s", "Bucket name invalid: .test")},
		{"------", "obj", "", 1, "", "", 0, false, "", fmt.Errorf("%s", "Bucket name invalid: ------")},
		{"$this-is-not-valid-too", "obj", "", 1, "", "", 0, false, "",
			fmt.Errorf("%s", "Bucket name invalid: $this-is-not-valid-too")},
		{"a", "obj", "", 1, "", "", 0, false, "", fmt.Errorf("%s", "Bucket name invalid: a")},
		// Test case - 5.
		// Case with invalid object names.
		{"abc", "", "", 1, "", "", 0, false, "", fmt.Errorf("%s", "Object name invalid: abc#")},
		// Test case - 6.
		// Valid object and bucket names but non-existent bucket.
		{"abc", "def", "", 1, "", "", 0, false, "", fmt.Errorf("%s", "Bucket not found: abc")},
		// Test Case - 7.
		// Existing bucket, but using a bucket on which NewMultipartUpload is not Initiated.
		{"unused-bucket", "def", "xyz", 1, "", "", 0, false, "", fmt.Errorf("%s", "Invalid upload id xyz")},
		// Test Case - 8.
		// Existing bucket, object name different from which NewMultipartUpload is constructed from.
		// Expecting "Invalid upload id".
		{bucket, "def", "xyz", 1, "", "", 0, false, "", fmt.Errorf("%s", "Invalid upload id xyz")},
		// Test Case - 9.
		// Existing bucket, bucket and object name are the ones from which NewMultipartUpload is constructed from.
		// But the uploadID is invalid.
		// Expecting "Invalid upload id".
		{bucket, object, "xyz", 1, "", "", 0, false, "", fmt.Errorf("%s", "Invalid upload id xyz")},
		// Test Case - 10.
		// Case with valid UploadID, existing bucket name.
		// But using the bucket name from which NewMultipartUpload is not constructed from.
		{"unused-bucket", object, uploadID, 1, "", "", 0, false, "", fmt.Errorf("%s", "Invalid upload id "+uploadID)},
		// Test Case - 10.
		// Case with valid UploadID, existing bucket name.
		// But using the object name from which NewMultipartUpload is not constructed from.
		{bucket, "none-object", uploadID, 1, "", "", 0, false, "", fmt.Errorf("%s", "Invalid upload id "+uploadID)},
		// Test case - 11.
		// Input to replicate Md5 mismatch.
		{bucket, object, uploadID, 1, "", "a35", 0, false, "",
			fmt.Errorf("%s", "Bad digest: Expected a35 is not valid with what we calculated "+"d41d8cd98f00b204e9800998ecf8427e")},
		// Test case - 12.
		// Input with size more than the size of actual data inside the reader.
		{bucket, object, uploadID, 1, "abcd", "a35", int64(len("abcd") + 1), false, "", fmt.Errorf("%s", "EOF")},
		// Test case - 13.
		// Input with size less than the size of actual data inside the reader.
		{bucket, object, uploadID, 1, "abcd", "a35", int64(len("abcd") - 1), false, "",
			fmt.Errorf("%s", "Contains more data than specified size of 3 bytes.")},
		// Test case - 14-17.
		// Validating for success cases.
		{bucket, object, uploadID, 1, "abcd", "e2fc714c4727ee9395f324cd2e7f331f", int64(len("abcd")), true, "", nil},
		{bucket, object, uploadID, 2, "efgh", "1f7690ebdd9b4caf8fab49ca1757bf27", int64(len("efgh")), true, "", nil},
		{bucket, object, uploadID, 3, "ijkl", "09a0877d04abf8759f99adec02baf579", int64(len("abcd")), true, "", nil},
		{bucket, object, uploadID, 4, "mnop", "e132e96a5ddad6da8b07bba6f6131fef", int64(len("abcd")), true, "", nil},
	}

	for i, testCase := range failCases {
		actualMd5Hex, actualErr := obj.PutObjectPart(testCase.bucketName, testCase.objName, testCase.uploadID, testCase.PartID, testCase.intputDataSize,
			bytes.NewBufferString(testCase.inputReaderData), testCase.inputMd5)
		// All are test cases above are expected to fail.

		if actualErr != nil && testCase.shouldPass {
			t.Errorf("Test %d: Expected to pass, but failed with: <ERROR> %s.", i+1, actualErr.Error())
		}
		if actualErr == nil && !testCase.shouldPass {
			t.Errorf("Test %d: Expected to fail with <ERROR> \"%s\", but passed instead.", i+1, testCase.expectedError.Error())
		}
		// Failed as expected, but does it fail for the expected reason.
		if actualErr != nil && !testCase.shouldPass {
			if testCase.expectedError.Error() != actualErr.Error() {
				t.Errorf("Test %d: Expected to fail with error \"%s\", but instead failed with error \"%s\" instead.", i+1,
					testCase.expectedError.Error(), actualErr.Error())
			}
		}
		// Test passes as expected, but the output values are verified for correctness here.
		if actualErr == nil && testCase.shouldPass {
			// Asserting whether the md5 output is correct.
			if testCase.inputMd5 != actualMd5Hex {
				t.Errorf("Test %d: Calculated Md5 different from the actual one %s.", i+1, actualMd5Hex)
			}
		}
	}
}
