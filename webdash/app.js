var angular = require('angular')
var angular_route = require('angular-route')
var ngtable = require('ng-table')
var mdl = require('material-design-lite')
var app = angular.module('TESApp', ['ngRoute', 'ngTable']);

function shortID(longID) {
  return longID.split('-')[0];
}

function listAllTasks($http, tasks, page) {
  if (!tasks) {
    tasks = []
  }

  var url = "/v1/tasks?view=BASIC";
  if (page) {
    url += "&page_token=" + page;
  }

  return $http.get(url).then(function(response) {
    Array.prototype.push.apply(tasks, response.data.tasks);
    if (response.data.next_page_token) {
      return listAllTasks($http, tasks, response.data.next_page_token);
    } else {
      return tasks
    }
  });
}

app.controller('TaskListController', function($scope, NgTableParams, $http, $interval) {
  $scope.shortID = shortID;
  var tasks = [];

  $scope.tableParams = new NgTableParams(
    {
      count: 100,
      sorting: { id: "desc" },
    }, 
    {
      counts: [25, 50, 100],
      paginationMinBlocks: 2,
      paginationMaxBlocks: 10,
      dataset: tasks,
    }
  );

  $scope.cancelTask = function(taskID) {
    var url = "/v1/tasks/" + taskID + ":cancel";
    $http.post(url);
  }

  function refresh() {
    listAllTasks($http).then(function(ts) {
      tasks.length = 0;
      Array.prototype.push.apply(tasks, ts);
      $scope.tableParams.total(tasks.length);
      $scope.tableParams.reload();
    });
  }
  refresh();
  stop = $interval(refresh, 2000);

  $scope.$on('$destroy', function() {
    $interval.cancel(stop);
  });
});

app.controller('NodeListController', function($scope, NgTableParams, $http, $interval) {

	$scope.url = "/v1/nodes";
	var nodes = [];

  $scope.tableParams = new NgTableParams(
    {
      count: 25,
      sorting: { state: "asc" },
    }, 
    {
      counts: [25, 50, 100],
      paginationMinBlocks: 2,
      paginationMaxBlocks: 10,
      dataset: nodes,
    }
  );

  function refresh() {
    $http.get($scope.url).then(function(response) {
      nodes.length = 0;
      Array.prototype.push.apply(nodes, response.data.nodes);
      $scope.tableParams.total(nodes.length);
      $scope.tableParams.reload();
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

app.controller('TaskInfoController', function($scope, $http, $routeParams, $location, $interval) {
  
  $scope.url = "/v1/tasks/" + $routeParams.task_id
  $scope.task = {};
  $scope.cmdStr = function(cmd) {
    return cmd.join(' ');
  };
  $scope.serverURL = getServerURL($location)

  $scope.cancelTask = function() {
    $http.post($scope.url + ":cancel");
  }

  function refresh() {
    $http.get($scope.url + "?view=FULL")
      .success(function(data, status, headers, config) {
        console.log(data);
        $scope.task = data;
      })
      .error(function(data, status, headers, config){
        console.log(data);
        $location.url("/notfound");
      });
  }
  refresh();
  stop = $interval(refresh, 2000);

  $scope.$on('$destroy', function() {
    $interval.cancel(stop);
  });
});

app.controller('NodeInfoController', function($scope, $http, $routeParams, $location, $interval) {
  
  $scope.url = "/v1/nodes/" + $routeParams.node_id
  $scope.node = {};
  $scope.serverURL = getServerURL($location)

  function refresh() {
    $http.get($scope.url)
      .success(function(data, status, headers, config) {
        console.log(data);
        $scope.node = data;
      })
      .error(function(data, status, headers, config) {
        console.log(data);
        $location.url("/notfound");
      });
  }
  refresh();
  stop = $interval(refresh, 2000);

  $scope.$on('$destroy', function() {
    $interval.cancel(stop);
  });
});

app.controller('Error404Controller', function() {});

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
       }).
       when('/v1/tasks/:task_id', {
         redirectTo: '/tasks/:task_id',
       }).
       when('/tasks/:task_id', {
         templateUrl: 'static/task.html'
       }).
       when('/v1/nodes', {
         redirectTo: '/nodes',
       }).
       when('/nodes', {
         templateUrl: 'static/node-list.html'
       }).
       when('/v1/nodes/:node_id', {
         redirectTo: '/nodes/:node_id',
       }).
       when('/nodes/:node_id', {
         templateUrl: 'static/node.html'
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
