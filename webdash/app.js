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

app.controller('TaskListController', function($scope, $http, $interval, $routeParams, $location) {
  $scope.pageTitle = "Tasks";
  $scope.tasks = [];
  $scope.isDone = isDone;
  $scope.serverURL = getServerURL($location)
  var page = $routeParams.page_token;

  $scope.cancelTask = function(taskID) {
    var url = "/v1/tasks/" + taskID + ":cancel";
    $http.post(url);
  }

  function refresh() {
    var url = "/v1/tasks?view=BASIC";
    if (page) {
      url += "&page_token=" + page;
    }

    $http.get(url).then(function(response) {
      $scope.$applyAsync(function() {
        $scope.tasks = response.data.tasks;
        $scope.tasks.sort(idDesc);
        if (response.data.next_page_token) {
          $scope.nextPage = $scope.serverURL + "/v1/tasks?page_token=" + response.data.next_page_token;
        } else {
          $scope.nextPage = "";
        }
      });
    });
  }
  refresh();
  stop = $interval(refresh, 2000);

  $scope.$on('$destroy', function() {
    $interval.cancel(stop);
  });
});

app.controller('NodeListController', function($scope, $http, $interval) {

	$scope.url = "/v1/nodes";
  $scope.nodes = [];

  function refresh() {
    $http.get($scope.url).then(function(response) {
      $scope.$applyAsync(function() {
        $scope.nodes = response.data.nodes;
        $scope.nodes.sort(idDesc);
      });
    });
  }
  refresh();
  stop = $interval(refresh, 2000);

  $scope.$on('$destroy', function() {
    $interval.cancel(stop);
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

app.controller('TaskInfoController', function($scope, $http, $routeParams, $location, $interval, Page) {
  
  $scope.url = "/v1/tasks/" + $routeParams.task_id;
  Page.setTitle("Task " + $routeParams.task_id);
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
  

  $scope.cancelTask = function() {
    $http.post($scope.url + ":cancel");
  }

  function refresh() {
    $http.get($scope.url + "?view=FULL")
    .success(function(data, status, headers, config) {
      $scope.task = data;
      $scope.loaded = true;
    })
    .error(function(data, status, headers, config){
      if (status == 404) {
        $scope.notFound = true;
        $interval.cancel(stop);
      }
    });
  }
  refresh();
  stop = $interval(refresh, 2000);

  $scope.$on('$destroy', function() {
    $interval.cancel(stop);
  });
});

app.controller('NodeInfoController', function($scope, $http, $routeParams, $location, $interval, Page, $filter) {
  
  $scope.url = "/v1/nodes/" + $routeParams.node_id;
  Page.setTitle("Node " + $routeParams.node_id);
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
    })
    .error(function(data, status, headers, config){
      if (status == 404) {
        $scope.notFound = true;
        $interval.cancel(stop);
      }
    });
  }
  refresh();
  stop = $interval(refresh, 2000);

  $scope.$on('$destroy', function() {
    $interval.cancel(stop);
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
    }
  });
}]);

app.config(
  ['$routeProvider', '$locationProvider',
   function AppConfig($routeProvider, $locationProvider) {
     var taskList = {
       templateUrl: '/static/list.html',
       title: "Tasks",
     }
     var taskInfo = {
       templateUrl: '/static/task.html',
     }
     var nodeList = {
       templateUrl: '/static/node-list.html',
       title: "Nodes",
     }
     var nodeInfo =  {
       templateUrl: '/static/node.html',
     }

     $routeProvider.
       when('/', taskList).
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
