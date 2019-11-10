package route

import (
    "github.com/gorilla/mux"
    asana "github.com/oms-services/asana/asana"
    "net/http"
)

type Route struct {
    Name        string
    Method      string
    Pattern     string
    HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
    Route{
        "CreateProject",
        "POST",
        "/createproject",
        asana.CreateProject,
    },
    Route{
        "CreateTask",
        "POST",
        "/createtask",
        asana.CreateTask,
    },
    Route{
        "DeleteProject",
        "POST",
        "/deleteproject",
        asana.DeleteProject,
    },
    Route{
        "DeleteTask",
        "POST",
        "/deletetask",
        asana.DeleteTask,
    },
    Route{
        "ListTask",
        "POST",
        "/listtask",
        asana.ListTask,
    },
    Route{
        "ListWorkspace",
        "POST",
        "/listworkspace",
        asana.ListWorkspace,
    },
    Route{
        "FindTask",
        "POST",
        "/findtask",
        asana.FindTask,
    },
    Route{
        "FindProject",
        "POST",
        "/findproject",
        asana.FindProject,
    },
    Route{
        "UpdateProject",
        "POST",
        "/updateproject",
        asana.UpdateProject,
    },
    Route{
        "UpdateProject",
        "POST",
        "/updateproject",
        asana.UpdateProject,
    },
    Route{
        "ListProjectTasks",
        "POST",
        "/listprojecttasks",
        asana.ListProjectTasks,
    },
    Route{
        "SubscribeTasks",
        "POST",
        "/subscribe",
        asana.SubscribeTasks,
    },
}

func NewRouter() *mux.Router {
    router := mux.NewRouter().StrictSlash(true)
    for _, route := range routes {
        var handler http.Handler
        handler = route.HandlerFunc

        router.
            Methods(route.Method).
            Path(route.Pattern).
            Name(route.Name).
            Handler(handler)
    }
    return router
}
