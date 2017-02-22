var app = angular.module('TESApp', ['ngRoute', 'ngTable']);

function shortID(longID) {
  return longID.split('-')[0];
}

app.controller('JobListController', function($scope, NgTableParams, $http) {
  $scope.url = "/v1/jobs";
  $scope.shortID = shortID;

  $http.get($scope.url).then(function(result) {
	  var jobs = result.data.jobs;
    $scope.tableParams = new NgTableParams(
      {
        count: 25
      }, 
      {
        counts: [25, 50, 100],
        paginationMinBlocks: 2,
        paginationMaxBlocks: 10,
        total: jobs.length,
        dataset: jobs
      }
    );
  });

  $scope.cancelJob = function(jobID) {
    var url = "/v1/jobs/" + jobID;
    $http.delete(url);
  }
});

app.controller('WorkerListController', function($scope, $http) {

	$scope.url = "/v1/jobs-service";
	$scope.workers = [];

	$scope.fetchContent = function() {
		$http.get($scope.url).then(function(result) {
			$scope.workers = result.data;
		});
	}

	$scope.fetchContent();
});

app.controller('JobInfoController', function($scope, $http, $routeParams) {
  
  $scope.url = "/v1/jobs/" + $routeParams.job_id

  $scope.job = {};
  $scope.cmdStr = function(cmd) {
    return cmd.join(' ');
  };

  $scope.fetchContent = function() {
    $http.get($scope.url).then(function(result){
      console.log(result.data);
      $scope.job = result.data
    })
  }
  $scope.fetchContent();

  $scope.cancelJob = function() {
    $http.delete($scope.url);
  }
});

app.config(
  ['$routeProvider',
   function($routeProvider) {
     $routeProvider.when('/', {
       templateUrl: 'static/list.html',
     }).
       when('/jobs/:job_id', {
         templateUrl: 'static/job.html'
       })
   }
  ]
);
