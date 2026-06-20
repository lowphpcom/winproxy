package main

import "testing"

func TestStatusTranslations(t *testing.T) {
	if tr("zh-CN").Stopped != "未启动" || tr("zh-CN").Started != "已启动" {
		t.Fatalf("unexpected zh-CN status translations")
	}
	if tr("en-US").Stopped != "Stopped" || tr("en-US").Started != "Running" {
		t.Fatalf("unexpected en-US status translations")
	}
}
