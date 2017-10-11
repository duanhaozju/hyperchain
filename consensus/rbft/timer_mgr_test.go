package rbft

//func TestPbftTimeFunctions(t *testing.T)  {
//	pp := new (pbftImpl)
//	pp.batchTimeout = 1 * time.Second
//	pp.batchTimerActive = false;
//	pp.batchTimer = events.NewTimerFactoryImpl(events.NewManagerImpl()).CreateTimer()
//
//	pp.startBatchTimer()
//	if(pp.batchTimeout != 1 * time.Second || pp.batchTimerActive == false){
//		t.Errorf("pbftProtocal %s not work!", "startBatchTimer")
//	}
//
//	pp.stopBatchTimer()
//	if pp.batchTimerActive == true {
//		t.Errorf("pbftProtocal %s not work!", "stopBatchTimer")
//	}
//
//	pp.newViewTimer =  events.NewTimerFactoryImpl(events.NewManagerImpl()).CreateTimer()
//	pp.startNewViewTimer(2 * time.Second, "test pbftProtocol viewTimer")
//	if pp.timerActive == false {
//		t.Errorf("pbftProtocal %s not work!", "startTimer")
//	}
//
//	pp.stopNewViewTimer()
//	if pp.timerActive == true {
//		t.Errorf("pbftProtocal %s not work!", "stopTimer")
//	}
//	rs := "test pbftProtocol softSDtartTimer"
//	pp.softStartTimer(1 * time.Second, rs)
//	if pp.timerActive == false || strings.Compare(pp.newViewTimerReason, rs) != 0 {
//		t.Errorf(`pbftProtocal %s not work!`, "softStartTimer")
//	}
//
//	/*pp.nullRequestTimer = events.NewTimerFactoryImpl(events.NewManagerImpl()).CreateTimer()
//	pp.nullRequestTimeout = 2 * time.Second
//	pp.id = 1
//	pp.view = 1
//	pp.requestTimeout = 3
//
//
//	pp.nullReqTimerReset()
//	pp.nullRequestTimer*/
//}
