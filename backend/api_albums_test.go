// API Test

package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
)

func Albums_API_Test_Create_Album(server *httptest.Server, session string, t *testing.T, name string) (albumId uint64, e error) {
	// Create album
	body, err := json.Marshal(CreateAlbumAPIBody{
		Name: name,
	})

	statusCode, bodyResponseBytes, err := DoTestRequest(server, "POST", "/api/albums", body, session)

	if err != nil {
		t.Error(err)
		return 0, err
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	res1 := CreateAlbumAPIResponse{}

	err = json.Unmarshal(bodyResponseBytes, &res1)

	if err != nil {
		t.Error(err)
		return 0, err
	}

	// Check

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/albums/"+fmt.Sprint(res1.Id), nil, session)

	if err != nil {
		t.Error(err)
		return 0, err
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	res2 := AlbumAPIDetail{}

	err = json.Unmarshal(bodyResponseBytes, &res2)

	if err != nil {
		t.Error(err)
		return 0, err
	}

	if res2.Name != name {
		t.Error(ErrorMismatch("AlbumName", res2.Name, name))
	}

	return res1.Id, nil
}

func Albums_API_Test(server *httptest.Server, session string, t *testing.T) {
	// Create albums

	album1Name := "Album 1"
	album2Name := "Album 2"

	album1, err := Albums_API_Test_Create_Album(server, session, t, album1Name)

	if err != nil {
		t.Error(err)
		return
	}

	album2, err := Albums_API_Test_Create_Album(server, session, t, album2Name)

	if err != nil {
		t.Error(err)
		return
	}

	// Get albums

	statusCode, bodyResponseBytes, err := DoTestRequest(server, "GET", "/api/albums?mode=min", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	res1 := make([]AlbumAPIItemMinified, 0)

	err = json.Unmarshal(bodyResponseBytes, &res1)

	if err != nil {
		t.Error(err)
		return
	}

	containsAlbum1 := false
	containsAlbum2 := false

	for i := 0; i < len(res1); i++ {
		if res1[i].Id == album1 {
			containsAlbum1 = true
			if res1[i].Name != album1Name {
				t.Error(ErrorMismatch("AlbumName", res1[i].Name, album1Name))
			}
		} else if res1[i].Id == album2 {
			containsAlbum2 = true
			if res1[i].Name != album2Name {
				t.Error(ErrorMismatch("AlbumName", res1[i].Name, album2Name))
			}
		}
	}

	if !containsAlbum1 {
		t.Errorf("Album list does not contain album 1")
	}

	if !containsAlbum2 {
		t.Errorf("Album list does not contain album 2")
	}

	// Rename album

	album2Name = "Album 2 - Renamed"

	body, err := json.Marshal(RenameAlbumAPIBody{
		Name: album2Name,
	})

	statusCode, _, err = DoTestRequest(server, "POST", "/api/albums/"+fmt.Sprint(album2)+"/rename", body, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/albums/"+fmt.Sprint(album2), nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	res2 := AlbumAPIDetail{}

	err = json.Unmarshal(bodyResponseBytes, &res2)

	if err != nil {
		t.Error(err)
		return
	}

	if res2.Name != album2Name {
		t.Error(ErrorMismatch("AlbumName", res2.Name, album2Name))
	}

	// Upload test media

	media1, err := UploadTestMedia(server, session, MediaTypeImage, "Test Media 1", fmt.Sprint(album1))

	if err != nil {
		t.Error(err)
		return
	}

	media2, err := UploadTestMedia(server, session, MediaTypeImage, "Test Media 2", fmt.Sprint(album1))

	if err != nil {
		t.Error(err)
		return
	}

	media3, err := UploadTestMedia(server, session, MediaTypeImage, "Test Media 3", fmt.Sprint(album2))

	if err != nil {
		t.Error(err)
		return
	}

	// Check albums

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/albums/"+fmt.Sprint(album1), nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	err = json.Unmarshal(bodyResponseBytes, &res2)

	if err != nil {
		t.Error(err)
		return
	}

	exceptedList := []uint64{media1, media2}

	if len(res2.List) != len(exceptedList) {
		t.Error(ErrorMismatch("AlbumListLength", fmt.Sprint(len(res2.List)), fmt.Sprint(len(exceptedList))))
	}

	for i := 0; i < len(res2.List); i++ {
		if res2.List[i].Id != exceptedList[i] {
			t.Error(ErrorMismatch("List["+fmt.Sprint(i)+"]", fmt.Sprint(res2.List[i].Id), fmt.Sprint(exceptedList[i])))
		}
	}

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/albums/"+fmt.Sprint(album2), nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	err = json.Unmarshal(bodyResponseBytes, &res2)

	if err != nil {
		t.Error(err)
		return
	}

	exceptedList = []uint64{media3}

	if len(res2.List) != len(exceptedList) {
		t.Error(ErrorMismatch("AlbumListLength", fmt.Sprint(len(res2.List)), fmt.Sprint(len(exceptedList))))
	}

	for i := 0; i < len(res2.List); i++ {
		if res2.List[i].Id != exceptedList[i] {
			t.Error(ErrorMismatch("List["+fmt.Sprint(i)+"]", fmt.Sprint(res2.List[i].Id), fmt.Sprint(exceptedList[i])))
		}
	}

	// Add media

	body, err = json.Marshal(AlbumMediaAPIBody{
		Id: media2,
	})

	statusCode, _, err = DoTestRequest(server, "POST", "/api/albums/"+fmt.Sprint(album2)+"/add", body, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/albums/"+fmt.Sprint(album2), nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	err = json.Unmarshal(bodyResponseBytes, &res2)

	if err != nil {
		t.Error(err)
		return
	}

	exceptedList = []uint64{media3, media2}

	if len(res2.List) != len(exceptedList) {
		t.Error(ErrorMismatch("AlbumListLength", fmt.Sprint(len(res2.List)), fmt.Sprint(len(exceptedList))))
	}

	for i := 0; i < len(res2.List); i++ {
		if res2.List[i].Id != exceptedList[i] {
			t.Error(ErrorMismatch("List["+fmt.Sprint(i)+"]", fmt.Sprint(res2.List[i].Id), fmt.Sprint(exceptedList[i])))
		}
	}

	// Remove media

	body, err = json.Marshal(AlbumMediaAPIBody{
		Id: media3,
	})

	statusCode, _, err = DoTestRequest(server, "POST", "/api/albums/"+fmt.Sprint(album2)+"/remove", body, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/albums/"+fmt.Sprint(album2), nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	err = json.Unmarshal(bodyResponseBytes, &res2)

	if err != nil {
		t.Error(err)
		return
	}

	exceptedList = []uint64{media2}

	if len(res2.List) != len(exceptedList) {
		t.Error(ErrorMismatch("AlbumListLength", fmt.Sprint(len(res2.List)), fmt.Sprint(len(exceptedList))))
	}

	for i := 0; i < len(res2.List); i++ {
		if res2.List[i].Id != exceptedList[i] {
			t.Error(ErrorMismatch("List["+fmt.Sprint(i)+"]", fmt.Sprint(res2.List[i].Id), fmt.Sprint(exceptedList[i])))
		}
	}

	// Set list

	body, err = json.Marshal(AlbumSetListAPIBody{
		List: []uint64{media3, media2, media1},
	})

	statusCode, _, err = DoTestRequest(server, "POST", "/api/albums/"+fmt.Sprint(album2)+"/set", body, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/albums/"+fmt.Sprint(album2), nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	err = json.Unmarshal(bodyResponseBytes, &res2)

	if err != nil {
		t.Error(err)
		return
	}

	exceptedList = []uint64{media3, media2, media1}

	if len(res2.List) != len(exceptedList) {
		t.Error(ErrorMismatch("AlbumListLength", fmt.Sprint(len(res2.List)), fmt.Sprint(len(exceptedList))))
	}

	for i := 0; i < len(res2.List); i++ {
		if res2.List[i].Id != exceptedList[i] {
			t.Error(ErrorMismatch("List["+fmt.Sprint(i)+"]", fmt.Sprint(res2.List[i].Id), fmt.Sprint(exceptedList[i])))
		}
	}

	// Delete album

	statusCode, _, err = DoTestRequest(server, "POST", "/api/albums/"+fmt.Sprint(album2)+"/delete", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	statusCode, _, err = DoTestRequest(server, "GET", "/api/albums/"+fmt.Sprint(album2), nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 404 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "404"))
	}

	// Check album list again

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/albums", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	res3 := make([]AlbumAPIItem, 0)

	err = json.Unmarshal(bodyResponseBytes, &res3)

	if err != nil {
		t.Error(err)
		return
	}

	containsAlbum1 = false
	containsAlbum2 = false

	for i := 0; i < len(res3); i++ {
		if res3[i].Id == album1 {
			containsAlbum1 = true

			if res3[i].Size != 2 {
				t.Error(ErrorMismatch("AlbumSize", fmt.Sprint(res3[i].Size), fmt.Sprint(2)))
			}

		} else if res3[i].Id == album2 {
			containsAlbum2 = true
		}
	}

	if !containsAlbum1 {
		t.Errorf("Album list does not contain album 1")
	}

	if containsAlbum2 {
		t.Errorf("Album list contains album 2 (deleted)")
	}
}
