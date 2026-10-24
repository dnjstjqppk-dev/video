package main

import "testing"

func TestLiveControllerLifecycleAndDVRHandoff(t *testing.T) {
	ctrl := NewLiveController()
	sess := ctrl.ProvisionInput("ls-101", "ch-1", true)
	if sess.State != LiveStateCreated || sess.StreamKey == "" {
		t.Fatalf("expected CREATED state and non-empty stream key")
	}

	oldKey := sess.StreamKey
	newKey, err := ctrl.RotateStreamKey("ls-101")
	if err != nil || newKey == oldKey {
		t.Fatalf("expected rotated stream key, err=%v", err)
	}

	if err := ctrl.TransitionState("ls-101", LiveStateLive); err != nil {
		t.Fatalf("expected CREATED -> LIVE transition to succeed: %v", err)
	}

	health := EvaluateHealth(6000, 60, 0.2)
	if health.Status != "HEALTHY" {
		t.Fatalf("expected HEALTHY status, got %s", health.Status)
	}

	if err := ctrl.TransitionState("ls-101", LiveStateEnded); err != nil {
		t.Fatalf("expected LIVE -> ENDED transition to succeed: %v", err)
	}
	if sess.VodVideoID != "vod_ls-101" {
		t.Fatalf("expected DVR recording handoff to VOD, got %s", sess.VodVideoID)
	}
}
