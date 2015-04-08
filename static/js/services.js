var bomServices = angular.module('bomServices', ['ngResource']);

bomServices.factory('BOM', ['$resource', function($resource){
    return $resource('/api/bom', {}, {
      'reset': {method: 'POST', url: '/api/bom/reset'},
      'resetToSampleState': {method: 'POST', url: '/api/bom/resetToSampleBOM'}
    });
  }]);

bomServices.factory('Model', ['$resource', function($resource){
    return $resource('/api/bom/model/:id',{},{
      'query':{method:'GET', isArrya:false}
    });
  }]);

bomServices.factory('Article', ['$resource', function($resource){
    return $resource('/api/bom/article/:id' );
  }]);

bomServices.factory('RRKSaleInvoice', ['$resource', function($resource){
    return $resource('/api/rrk/saleInvoice/:id',{},{
    } );
  }]);
