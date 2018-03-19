var angular = require('angular')
var angular_route = require('angular-route')
var mdl = require('material-design-lite')
var app = angular.module('TESApp', ['ngRoute']);

function idDesc(a, b) {
  if (a.id == b.id) {
    return 0;
  } 
  if (a.id < b.id) {
    return 1; 
  }
  return -1;
}

function isDone(task) {
  return task.state == 'COMPLETE' || task.state == 'EXECUTOR_ERROR' || task.state == 'CANCELED' || task.state == 'SYSTEM_ERROR';
}

app.service("TaskFilters", function($rootScope) {
  var s = $rootScope.$new()
  s.state = "any";
  s.tags = [];
  return s;
})

app.controller('TaskFilterController', function($scope, TaskFilters) {
  $scope.filters = TaskFilters;

  $scope.addNewTag = function(tag) {
    $scope.filters.tags.push({
      'key': "",
      'value': "",
    });
  };

  $scope.removeTag = function(tag) {
    index = $scope.filters.tags.indexOf(tag);
    $scope.filters.tags.splice(index, 1);
  };
})

app.controller('TaskListController', function($rootScope, $scope, $http, $timeout, $routeParams, $location, TaskFilters) {
  $rootScope.pageTitle = "Tasks";
  $scope.tasks = [];
  $scope.isDone = isDone;
  $scope.serverURL = getServerURL($location)
  var page = $routeParams.page_token;

  $scope.cancelTask = function(taskID) {
    var url = "/v1/tasks/" + taskID + ":cancel";
    $http.post(url);
  }

  TaskFilters.$watch("state", function() {
    refresh();
  })

  TaskFilters.$watch("tags", function() {
    refresh();
  })

  function listTasks() {
    var url = "/v1/tasks?view=BASIC";
    if (page) {
      url += "&page_token=" + page;
    }
    if (TaskFilters.state != "any") {
      url += "&state=" + TaskFilters.state;
    }
    if (TaskFilters.tags.length) {
      for (i in TaskFilters.tags) {
        tag = TaskFilters.tags[i];
        if (tag.key) {
          url += "&tags%5B"+tag.key+"%5D"
          if (tag.value) {
            url += "="+tag.value
          }
        }
      }
    }
    return $http.get(url);
  }

  function refresh() {
    listTasks().then(function(response) {
      $scope.$applyAsync(function() {
        $scope.tasks = response.data.tasks;
        if (response.data.next_page_token) {
          $scope.nextPage = $scope.serverURL + "/v1/tasks?page_token=" + response.data.next_page_token;
        } else {
          $scope.nextPage = "";
        }
      });
    });
  }

  function autoRefresh() {
    listTasks().then(function(response) {
      $scope.$applyAsync(function() {
        $scope.tasks = response.data.tasks;
        if (response.data.next_page_token) {
          $scope.nextPage = $scope.serverURL + "/v1/tasks?page_token=" + response.data.next_page_token;
        } else {
          $scope.nextPage = "";
        }
        stop = $timeout(autoRefresh, 2000);
      });
    });
  }

  autoRefresh();

  $scope.$on('$destroy', function() {
    $timeout.cancel(stop);
  });
});

app.controller('NodeListController', function($rootScope, $scope, $http, $timeout) {
  $rootScope.pageTitle = "Nodes";
	$scope.url = "/v1/nodes";
  $scope.nodes = [];

  function refresh() {
    $http.get($scope.url).then(function(response) {
      $scope.$applyAsync(function() {
        $scope.nodes = response.data.nodes;
        $scope.nodes.sort(idDesc);
        stop = $timeout(refresh, 2000);
      });
    });
  }

  refresh();

  $scope.$on('$destroy', function() {
    $timeout.cancel(stop);
  });
});

function getServerURL($location) {
  var proto = $location.protocol();
  var port = $location.port();

  // If the port is a standard HTTP(S) port, don't include it in the URL.
  if ((proto == "https" && port == 443) || (proto == "http" && port == 80)) {
    return proto + "://" + $location.host();
  }

  return proto + "://" + $location.host() + ":" + port;
}

app.controller('TaskInfoController', function($rootScope, $scope, $http, $routeParams, $location, $timeout, Page) {
  $rootScope.pageTitle = "Task " + $routeParams.task_id;
  Page.setTitle("Task " + $routeParams.task_id);
  $scope.url = "/v1/tasks/" + $routeParams.task_id;
  $scope.task = {};
  $scope.cmdStr = function(cmd) {
    return cmd.join(' ');
  };
  $scope.serverURL = getServerURL($location)
  $scope.isDone = isDone;
  $scope.resources = function(task) {
    r = task.resources;
    if (angular.equals(r, {}) || r == undefined) {
      return "";
    }
    s = r.cpu_cores + " CPU cores";
    if (r.ram_gb) {
      s += ", " + r.ram_gb + " GB RAM";
    }
    if (r.disk_gb) {
      s += ", " + r.disk_gb + " GB disk space";
    }
    if (r.preemptible) {
      s += ", preemptible";
    }
    return s;
  }
  
  $scope.truncateContent = function(input) {
    if (input.content == "" || input.content == undefined) {
      return "";
    }
    if (input.content.length > 200) {
      return input.content.substring(0,200)+" ...";
    }
    return input.content;
  }

  $scope.cancelTask = function() {
    $http.post($scope.url + ":cancel");
  }

  $scope.syslogs = [];
  function parseSystemLogs(task) {
    if (!task || !task.logs || task.logs.length == 0 || !task.logs[0].system_logs) {
      return
    }
    $scope.syslogs = [];

    for (var i = 0; i < task.logs[0].system_logs.length; i++) {
      var log = task.logs[0].system_logs[i]
      var logre = /(\w+)='([^'\\]*(?:\\.[^'\\]*)*)'/g;

      var m;
      var parts = [];
      var level = "info";
      do {
          m = logre.exec(log);
          if (m) {
              var p = {key: m[1], value: m[2]};
              if (p.key == "level") {
                level = p.value
              }
              parts.push(p);
          }
      } while (m);

      if (parts.length > 0) {
        $scope.syslogs.push({level: level, parts: parts});
      }
    }
  }

  $scope.entryClass = function(entry) {
    return entry.level + "-level";
  }

  function refresh() {
    if (!$scope.isDone($scope.task)) {
      $http.get($scope.url + "?view=FULL")
        .success(function(data, status, headers, config) {
          $scope.task = data;
          parseSystemLogs(data);
          $scope.loaded = true;
          stop = $timeout(refresh, 2000);
        })
        .error(function(data, status, headers, config){
          if (status == 404) {
            $scope.notFound = true;
            $timeout.cancel(stop);
          }
        });
    }
  }

  refresh();

  $scope.$on('$destroy', function() {
    $timeout.cancel(stop);
  });
});

app.controller('NodeInfoController', function($rootScope, $scope, $http, $routeParams, $location, $timeout, $filter, Page) {
  $rootScope.pageTitle = "Node " + $routeParams.node_id;
  Page.setTitle("Node " + $routeParams.node_id);
  $scope.url = "/v1/nodes/" + $routeParams.node_id;
  $scope.node = {};
  $scope.serverURL = getServerURL($location)
  $scope.resources = function(r) {
    if (angular.equals(r, {}) || r == undefined) {
      return "";
    }
    s = r.cpus + " CPU cores";
    if (r.ram_gb) {
      s += ", " + $filter("number")(r.ram_gb) + " GB RAM";
    }
    if (r.disk_gb) {
      s += ", " + $filter("number")(r.disk_gb) + " GB disk space";
    }
    return s;
  }

  function refresh() {
    $http.get($scope.url)
    .success(function(data, status, headers, config) {
      $scope.node = data;
      $scope.loaded = true;
      stop = $timeout(refresh, 2000);
    })
    .error(function(data, status, headers, config){
      if (status == 404) {
        $scope.notFound = true;
        $timeout.cancel(stop);
      }
    });
  }

  refresh();

  $scope.$on('$destroy', function() {
    $timeout.cancel(stop);
  });
});

app.controller('ServiceInfoController', function($scope, $http, $location) {
  $http.get("/v1/tasks/service-info")
  .success(function(data, status, headers, config) {
    $scope.name = data.name;
    $scope.doc = data.doc;
  })
  .error(function(data, status, headers, config){
    $scope.error = data;
  });
});

app.controller('Error404Controller', function() {});

app.service('Page', function($rootScope){
  $rootScope.page = {
    title: "Funnel",
  }
  return {
    setTitle: function(title){
      $rootScope.page.title = title + " | Funnel";
    }
  }
});

app.run(['$rootScope', 'Page', function($rootScope, Page) {
  $rootScope.$on("$routeChangeSuccess", function(event, current, previous){
    if (current.$$route) {
      Page.setTitle(current.$$route.title);
      $rootScope.pageId = current.$$route.pageId;
    }
  });
}]);

app.config(
  ['$routeProvider', '$locationProvider',
   function AppConfig($routeProvider, $locationProvider) {
     var taskList = {
       templateUrl: '/static/list.html',
       title: "Tasks",
       pageId: "task-list",
     }
     var taskInfo = {
       templateUrl: '/static/task.html',
       pageId: "task-info",
     }
     var nodeList = {
       templateUrl: '/static/node-list.html',
       title: "Nodes",
       pageId: "node-list",
     }
     var nodeInfo =  {
       templateUrl: '/static/node.html',
       pageId: "node-info",
     }
     var serviceInfo = {
       templateUrl: "/static/service.html",
       title: "Service",
       pageId: "service-info",
     }

     $routeProvider.
       when('/', taskList).
       when('/v1/tasks/service-info', serviceInfo).
       when('/tasks/service-info', serviceInfo).
       when('/service-info', serviceInfo).
       when('/tasks', taskList).
       when('/v1/tasks', taskList).
       when('/tasks/:task_id', taskInfo).
       when('/v1/tasks/:task_id', taskInfo).
       when('/nodes', nodeList).
       when('/v1/nodes', nodeList).
       when('/nodes/:node_id', nodeInfo).
       when('/v1/nodes/:node_id', nodeInfo).
       otherwise({
         templateUrl: '/static/error404.html'
       });
     $locationProvider.html5Mode(true);
   }
  ]
);
