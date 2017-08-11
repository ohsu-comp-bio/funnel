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

  return $http.get(url).then(function(result) {
    Array.prototype.push.apply(tasks, result.data.tasks);
    if (result.data.next_page_token) {
      return listAllTasks($http, tasks, result.data.next_page_token);
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

  function refresh() {
    listAllTasks($http).then(function(ts) {
      tasks.length = 0;
      Array.prototype.push.apply(tasks, ts);
      $scope.tableParams.total(tasks.length);
      $scope.tableParams.reload();
    });
  }
  $interval(refresh, 2000);

  $scope.cancelTask = function(taskID) {
    var url = "/v1/tasks/" + taskID + ":cancel";
    $http.post(url);
  }
});

app.controller('WorkerListController', function($scope, $http) {

	$scope.url = "/v1/funnel/workers";
	$scope.workers = [];

  $http.get($scope.url).then(function(result) {
    var workers = result.data.workers || [];
console.log(workers)
    $scope.workers = workers;
  });
});

app.controller('TaskInfoController', function($scope, $http, $routeParams, $location, $interval) {
  
  $scope.url = "/v1/tasks/" + $routeParams.task_id + "?view=FULL"

  $scope.task = {};
  $scope.cmdStr = function(cmd) {
    return cmd.join(' ');
  };

  function refresh() {
    $http.get($scope.url).then(function(result){
      console.log(result.data);
      $scope.task = result.data
    })
  }
  refresh();
  $interval(refresh, 2000);

  $scope.serverURL = $location.protocol() + "://" + $location.host() + ":" + $location.port();

  $scope.cancelTask = function() {
    $http.post($scope.url + ":cancel");
  }
});

app.config(
  ['$routeProvider',
   function($routeProvider) {
     $routeProvider.when('/', {
       templateUrl: 'static/list.html',
     }).
     when('/tasks/:task_id', {
       templateUrl: 'static/task.html'
     }).
     when('/workers/', {
       templateUrl: 'static/worker-list.html'
     })
   }
  ]
);
