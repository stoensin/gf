// Copyright 2018 gf Author(https://gitee.com/johng/gf). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://gitee.com/johng/gf.

package gtcp

import (
    "net"
    "time"
    "gitee.com/johng/gf/g/container/gmap"
    "gitee.com/johng/gf/g/container/gpool"
)

// 封装的链接对象
type Conn struct {
    net.Conn         // 继承底层链接接口对象
    pool *gpool.Pool // 对应的链接池对象
    err  error       // 是否有错误产生
}

const (
    gDEFAULT_POOL_EXPIRE = 3000 // (毫秒)默认链接对象过期时间
)

var (
    // 连接池对象map，键名为地址端口，键值为对应的连接池对象
    pools = gmap.NewStringInterfaceMap()
)

// 创建TCP链接
func NewConn(addr string, timeout...int) (*Conn, error) {
    var pool *gpool.Pool
    if v := pools.Get(addr); v == nil {
        pools.LockFunc(func(m map[string]interface{}) {
            if v, ok := m[addr]; ok {
                pool = v.(*gpool.Pool)
            } else {
                pool = gpool.New(gDEFAULT_POOL_EXPIRE, func() (interface{}, error) {
                    if conn, err := NewNetConn(addr, timeout...); err == nil {
                        return &Conn {
                            Conn : conn,
                            pool : pool,
                        }, nil
                    } else {
                        return nil, err
                    }
                })
                m[addr] = pool
            }
        })
    } else {
        pool = v.(*gpool.Pool)
    }

    if v, err := pool.Get(); err == nil {
        return v.(*Conn), nil
    } else {
        return nil, err
    }
}

// 将net.Conn接口对象转换为*gtcp.Conn对象(注意递归影响，因为*gtcp.Conn本身也实现了net.Conn接口)
func NewConnByNetConn(conn net.Conn) *Conn {
    return &Conn {
        Conn : conn,
    }
}

// 覆盖底层接口对象的Close方法
func (c *Conn) Close() error {
    if c.pool != nil && c.err == nil {
        c.pool.Put(c)
    } else {
        c.Conn.Close()
    }
    return nil
}

// 发送数据
func (c *Conn) Send(data []byte, retry...Retry) error {
    err := Send(c, data, retry...)
    if err != nil {
        c.err = err
    }
    return err
}

// 接收数据
func (c *Conn) Receive(retry...Retry) ([]byte, error) {
    data, err := Receive(c, retry...)
    if err != nil {
        c.err = err
    }
    return data, err
}

// 带超时时间的数据获取
func (c *Conn) ReceiveWithTimeout(timeout time.Duration, retry...Retry) ([]byte, error) {
    data, err := ReceiveWithTimeout(c, timeout, retry...)
    if err != nil {
        c.err = err
    }
    return data, err
}

// 带超时时间的数据发送
func (c *Conn) SendWithTimeout(data []byte, timeout time.Duration, retry...Retry) error {
    err := SendWithTimeout(c, data, timeout, retry...)
    if err != nil {
        c.err = err
    }
    return err
}

// 发送数据并等待接收返回数据
func (c *Conn) SendReceive(data []byte, retry...Retry) ([]byte, error) {
    data, err := SendReceive(c, data, retry...)
    if err != nil {
        c.err = err
    }
    return data, err
}

// 发送数据并等待接收返回数据(带返回超时等待时间)
func (c *Conn) SendReceiveWithTimeout(data []byte, timeout time.Duration, retry...Retry) ([]byte, error) {
    data, err := SendReceiveWithTimeout(c, data, timeout, retry...)
    if err != nil {
        c.err = err
    }
    return data, err
}