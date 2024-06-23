package user_mutex

import "sync"

type UserMutexT struct {
    locks sync.Map
}

// 获取用户 mutex
func (ul *UserMutexT) getLock(userID string) *sync.Mutex {
    lock, _ := ul.locks.LoadOrStore(userID, &sync.Mutex{})
    return lock.(*sync.Mutex)
}