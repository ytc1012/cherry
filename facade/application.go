package cherryFacade

type (
	IApplication interface {
		INode
		ISerializer
		IPacketCodec
		Running() bool                    // 是否运行中
		Find(name string) IComponent      // 根据name获取组件对象
		Remove(name string) IComponent    // 根据name移除组件对象
		All() []IComponent                // 获取所有组件列表
		OnShutdown(fn ...func())          // 关闭前执行的函数
		Startup(components ...IComponent) // 启动应用实例
		Shutdown()                        // 关闭应用实例
	}

	// IAppContext App上下文
	IAppContext interface {
		Set(app IApplication)
		App() IApplication
	}

	// AppContext 继承自IApplication实现默认的方法
	AppContext struct {
		app IApplication
	}
)

func (b *AppContext) Set(app IApplication) {
	b.app = app
}

func (b *AppContext) App() IApplication {
	return b.app
}