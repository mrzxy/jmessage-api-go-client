package jmessage

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/franela/goreq"
	"io/ioutil"
	"net/http"
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