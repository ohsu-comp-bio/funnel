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
  return task.state == 'COMPLETE' || task.state == 'ERROR' || task.state == 'CANCELED' || task.state == 'SYSTEM_ERROR';
}

app.controller('TaskListController', function($scope, $http, $interval, $routeParams, $location) {
  $scope.pageTitle = "Tasks";
  $scope.tasks = [];
  $scope.isDone = isDone;
  $scope.serverURL = getServerURL($location)
  var page = $routeParams.page;

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
          $scope.nextPage = $scope.serverURL + "/#/tasks?page_token=" + response.data.next_page_token;
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
  Page.setHeader("Task " + $routeParams.task_id);
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
    if (r.size_gb) {
      s += ", " + r.size_gb + " GB disk space";
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
    $http.get($scope.url + "?view=FULL").success(function(data, status, headers, config) {
      $scope.task = data;
    })
  }
  refresh();
  stop = $interval(refresh, 2000);

  $scope.$on('$destroy', function() {
    $interval.cancel(stop);
  });
});

app.controller('NodeInfoController', function($scope, $http, $routeParams, $location, $interval, Page, $filter) {
  
  $scope.url = "/v1/nodes/" + $routeParams.node_id;
  Page.setHeader("Node " + $routeParams.node_id);
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
    $http.get($scope.url).success(function(data, status, headers, config) {
      $scope.node = data;
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
    header: "",
  }
  return {
    setHeader: function(header){
      $rootScope.page.title = header + " | Funnel";
      $rootScope.page.header = header;
    }
  }
});

app.run(['$rootScope', 'Page', function($rootScope, Page) {
  $rootScope.$on("$routeChangeSuccess", function(event, current, previous){
    Page.setHeader(current.$$route.title || '');
  });
}]);

app.config(
  ['$routeProvider', '$locationProvider',
   function AppConfig($routeProvider, $locationProvider) {
     $routeProvider.
       when('/', {
         redirectTo: '/tasks',
       }).
       when('/v1/tasks', {
         redirectTo: '/tasks',
       }).
       when('/tasks', {
         templateUrl: 'static/list.html',
         title: "Tasks",
       }).
       when('/v1/tasks/:task_id', {
         redirectTo: '/tasks/:task_id',
       }).
       when('/tasks/:task_id', {
         templateUrl: 'static/task.html',
       }).
       when('/v1/nodes', {
         redirectTo: '/nodes',
       }).
       when('/nodes', {
         templateUrl: 'static/node-list.html',
         title: "Nodes",
       }).
       when('/v1/nodes/:node_id', {
         redirectTo: '/nodes/:node_id',
       }).
       when('/nodes/:node_id', {
         templateUrl: 'static/node.html',
       }).
       when('/notfound', {
         templateUrl: 'static/error404.html'
       }).
       otherwise({
         templateUrl: 'static/error404.html'
       });
     $locationProvider.html5Mode(false);
   }
  ]
);
