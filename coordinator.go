package saga

import (
	"fmt"
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"reflect"
	"strconv"
)

// DefaultSEC is default SEC use by package method
var DefaultSEC ExecutionCoordinator = NewSEC()

// ExecutionCoordinator 代表 Saga 的协调器.
// 负责管理:
// - Saga log 存储.
// - 根据参数param信息定义短事务.
type ExecutionCoordinator struct {
	subTxDefinitions  subTxDefinitions
	paramTypeRegister *paramTypeRegister
}

// 创建 Saga 协调器
// 该方法需要一个日志存储
func NewSEC() ExecutionCoordinator {
	return ExecutionCoordinator{
		subTxDefinitions: make(subTxDefinitions),
		paramTypeRegister: &paramTypeRegister{
			nameToType: make(map[string]reflect.Type),
			typeToName: make(map[reflect.Type]string),
		},
	}
}

// AddSubTxDef
//	创建或添加基于给定短事务ID、动作、赔偿的定义，并且返回一个协调器
// 本方法是对默认协调器的操作
// subTxID 标示一个短事务的类型, 同时该id会被保存到日志，赔偿操作会用到
// action 短事务对应的服务会做的动作
// compensate 回滚时会做的动作
//
// action 和 compensat 两个参数必须是函数，而且函数的第一个参数必须是context.Context
func AddSubTxDef(subTxID string, action interface{}, compensate interface{}) *ExecutionCoordinator {
	return DefaultSEC.AddSubTxDef(subTxID, action, compensate)
}

// AddSubTxDef（方法）
//	创建或添加基于给定短事务ID、动作、赔偿的定义，并且返回一个协调器
// 本方法是对默认协调器的操作
// subTxID 标示一个短事务的类型, 同时该id会被保存到日志，赔偿操作会用到
// action 短事务对应的服务会做的动作
// compensate 回滚时会做的动作
//
// action 和 compensat 两个参数必须是函数，而且函数的第一个参数必须是context.Context
func (e *ExecutionCoordinator) AddSubTxDef(subTxID string, action interface{}, compensate interface{}) *ExecutionCoordinator {
	e.paramTypeRegister.addParams(action)
	e.paramTypeRegister.addParams(compensate)
	e.subTxDefinitions.addDefinition(subTxID, action, compensate)
	return e
}

// MustFindSubTxDef
// 根据给定的subTxID返回短事务
// 如果查找不到则报panic
func (e *ExecutionCoordinator) MustFindSubTxDef(subTxID string) subTxDefinition {
	define, ok := e.subTxDefinitions.findDefinition(subTxID)
	if !ok {
		panic("SubTxID: " + subTxID + " not found in context")
	}
	return define
}

// MustFindParamName
// 通过给定的反射类型返回参数名
// 参数无法找到则报错
func (e *ExecutionCoordinator) MustFindParamName(typ reflect.Type) string {
	name, ok := e.paramTypeRegister.findTypeName(typ)
	if !ok {
		panic("Find Param Name Panic: " + typ.String())
	}
	return name
}

// MustFindParamType
// 通过给定的参数名返回参数类型
// 如果找不到则报错
func (e *ExecutionCoordinator) MustFindParamType(name string) reflect.Type {
	typ, ok := e.paramTypeRegister.findType(name)
	if !ok {
		panic("Find Param Type Panic: " + name)
	}
	return typ
}

func (e *ExecutionCoordinator) StartCoordinator() error {
	logIDs, err := LogStorage().LogIDs()
	if err != nil {
		return errors.Annotate(err, "Fetch logs failure")
	}
	for _, logID := range logIDs {
		lastLogData, err := LogStorage().LastLog(logID)
		if err != nil {
			return errors.Annotate(err, "Fetch last log panic")
		}
		fmt.Println(lastLogData)
	}
	return nil
}

// StartSaga
// 启动一个新的saga，通过默认的协调器返回一个saga
// 参数需要一个context和一个唯一的id用于区分不同的saga
func StartSaga(ctx context.Context, id uint64) *Saga {
	return DefaultSEC.StartSaga(ctx, id)
}

// StartSaga
// 直接启动一个新的saga，并且返回
// 参数需要一个context和一个唯一的id用于区分不同的saga
func (e *ExecutionCoordinator) StartSaga(ctx context.Context, id uint64) *Saga {
	s := &Saga{
		id:      id,
		context: ctx,
		sec:     e,
		logID:   LogPrefix + strconv.FormatInt(int64(id), 10),
	}
	s.startSaga()
	return s
}
