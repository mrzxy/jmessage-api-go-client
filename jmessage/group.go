package jmessage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/franela/goreq"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type CreateGroupOpt struct {
	Name string `json:"name"`
	OwnerUsername string `json:"owner_username"`
	Avatar string `json:"avatar"`
	Desc string `json:"desc"`
	Flag int8 `json:"flag"`
	MembersUsername []string `json:"members_username"`
}
type CreateGroupRst struct {
	GID int `json:"gid"`
	OwnerUsername string `json:"owner_username"`
	Name string `json:"name"`
	MembersUsername []string `json:"members_username"`
	Desc string `json:"desc"`
	MaxMemberCount int `json:"max_member_count"`

}
func (jclient *JMessageClient) CreateGroup(opt CreateGroupOpt) (*CreateGroupRst, error) {
	req := goreq.Request{
		Method:            "POST",
		Uri:               JMESSAGE_IM_URL + GROUPS_URL,
		Accept:            "application/json",
		ContentType:       "application/json",
		UserAgent:         "JMessage-API-GO-Client",
		BasicAuthUsername: jclient.appKey,
		BasicAuthPassword: jclient.masterSecret,
		Timeout:           30 * time.Second, //30s
	}
	req.Body = opt
	req.ShowDebug = jclient.showDebug

	res, err := req.Do()

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	ibytes, err := ioutil.ReadAll(res.Body)
	if nil != err {
		return nil, err
	}
	if jclient.showDebug {
		fmt.Println("respone:", string(ibytes))
	}
	rst := &CreateGroupRst{}
	if err = json.Unmarshal(ibytes, rst); err == nil {
		return rst, nil
	}
	jRst := JMResponse{}
	if err = json.Unmarshal(ibytes, &jRst); err == nil {
		return nil, jRst.Error
	}

	return nil, errors.New("未知错误")
}

type UpdateGroupMemberOpt struct {
	GroupID int
	Members []string
}
func (jclient *JMessageClient) AddGroupMember(opt UpdateGroupMemberOpt) error {
	url := fmt.Sprintf("%s%s%d/addMembers", JMESSAGE_IM_URL, GROUPS_V2_URL, opt.GroupID)
	res, err := jclient.request(url, "POST", opt.Members)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		if err := jclient.handleJmErr(res.Body); err != nil {
			return err
		}
	}
	return nil
}

func (jclient *JMessageClient) RemoveGroupMember(opt UpdateGroupMemberOpt) error {
	url := fmt.Sprintf("%s%s%d/delMembers", JMESSAGE_IM_URL, GROUPS_V2_URL, opt.GroupID)
	res, err := jclient.request(url, "POST", opt.Members)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		if err := jclient.handleJmErr(res.Body); err != nil {
			return err
		}
	}
	return nil
}

// 移交群主
type ChangeGroupOwnerOpt struct {
	GroupID int
	Username string
}
func (jclient *JMessageClient) ChangeGroupOwner(opt ChangeGroupOwnerOpt) error {
	url := fmt.Sprintf("%s/groups/owner/%d", JMESSAGE_IM_URL, opt.GroupID)
	res, err := jclient.request(url, "PUT", map[string]string{"username": opt.Username})
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		if err := jclient.handleJmErr(res.Body); err != nil {
			return err
		}
	}
	return nil
}

type GetGroupMemberRst struct {
	Username string
	Nickname string
	Avatar string
	Birthday string
	Gender int8
	Signature string
	Region string
	Address string
	Flag int8
}
func (jclient *JMessageClient) GetGroupMember(groupID int) ([]GetGroupMemberRst, error) {
	rst := make([]GetGroupMemberRst, 0)
	url := fmt.Sprintf("%s%s%d/members/", JMESSAGE_IM_URL, GROUPS_URL, groupID)
	res, err := jclient.request(url, "GET", nil)
	if err != nil {
		return rst, err
	}
	defer res.Body.Close()

	ibytes, err := ioutil.ReadAll(res.Body)
	if nil != err {
		return rst, err
	}
	if jclient.showDebug {
		fmt.Println("respone:", string(ibytes))
	}
	jmErr := JMResponse{}
	if err := json.Unmarshal(ibytes, &jmErr); err == nil {
		return rst, errors.New(jmErr.Error.Message)
	}

	err = json.Unmarshal(ibytes, &rst)
	return rst, err
}

// 获取群组详情
type GetGroupInfoRst struct {
	GID int `json:"gid"`
	Name string `json:"name"`
	Desc string `json:"desc"`
	AppKey string `json:"appKey"`
	MaxMemberCount int `json:"max_member_count"`
	Mtime string
	Ctime string
}

func (jclient *JMessageClient) GetGroupInfo(groupID int) (*GetGroupInfoRst, error) {
	url := fmt.Sprintf("%s%s%d", JMESSAGE_IM_URL, GROUPS_URL, groupID)
	res, err := jclient.request(url, "GET", nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	rst := &GetGroupInfoRst{}
	if err := jclient.handleGetJmErr(res.Body, rst); err == nil {
		return rst, nil
	} else {
		return nil, err
	}
}

// 移交群主
type UpdateGroupOpt struct {
	GroupID int
	Name string
}
func (jclient *JMessageClient) UpdateGroup(opt UpdateGroupOpt) error {
	url := fmt.Sprintf("%s%s%d", JMESSAGE_IM_URL, GROUPS_URL, opt.GroupID)
	res, err := jclient.request(url, "PUT", map[string]string{"name": opt.Name})
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		if err := jclient.handleJmErr(res.Body); err != nil {
			return err
		}
	}
	return nil
}

// 上传媒体文件
func (jclient *JMessageClient) UploadMedia(imgUrl string) (*JPIMGMsg, error) {
	resp, err2 := http.Get(imgUrl)
	if err2 != nil {
		return nil, err2
	}
	defer resp.Body.Close()

	url := fmt.Sprintf("%s/v1/resource?type=image", JMESSAGE_IM_URL)
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	filename := imgUrl[strings.LastIndex(imgUrl, "/"):] + ".png"
	fileWriter, _ := bodyWriter.CreateFormFile("filename", filename)
	contentType := bodyWriter.FormDataContentType()
	if _, err := io.Copy(fileWriter, resp.Body); err != nil {
		return nil, err
	}
	bodyWriter.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bodyBuf)
	req.SetBasicAuth(jclient.appKey, jclient.masterSecret)
	req.Header.Set("Content-Type", contentType)
	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	rst := &JPIMGMsg{}
	if err := jclient.handleGetJmErr(res.Body, rst); err == nil {
		return rst, nil
	} else {
		return nil, err
	}
}

// 移交群主
type MuteUserByGroupOpt struct {
	GroupID int
	Status bool
	Members []string
}
func (jclient *JMessageClient) MuteUserByGroup(opt MuteUserByGroupOpt) error {
	url := fmt.Sprintf("%s/v1/groups/messages/%d/silence?status=%v", JMESSAGE_IM_URL, opt.GroupID, opt.Status)
	fmt.Println(url)
	res, err := jclient.request(url, "PUT", opt.Members)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		if err := jclient.handleJmErr(res.Body); err != nil {
			return err
		}
	}
	return nil
}