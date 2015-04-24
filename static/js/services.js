var bomServices = angular.module('bomServices', ['ngResource']);

bomServices.factory('BOM', ['$resource', function($resource){
    return $resource('/api/bom', {}, {
      'reset': {method: 'POST', url: '/api/bom/reset'},
      'resetToSampleState': {method: 'POST', url: '/api/bom/resetToSampleBOM'}
    });
  }]);

bomServices.factory('Model', ['$resource', function($resource){
    return $resource('/api/bom/model/:id',{},{
      'query':{method:'GET', isArray:false}
    });
  }]);

bomServices.factory('Article', ['$resource', function($resource){
    return $resource('/api/bom/article/:id',{},{
      'query':{method:'GET', isArray:false}
    });
  }]);

bomServices.factory('RRKSaleInvoice', ['$resource', function($resource){
    return $resource('/api/rrk/saleInvoice/:id',{},{
      'query':{method:'GET', isArray:false}
    } );
  }]);

bomServices.factory('RRKPurchaseInvoice', ['$resource', function($resource){
    return $resource('/api/rrk/purchaseInvoice/:id',{},{
      'query':{method:'GET', isArray:false}
    } );
  }]);

bomServices.factory('RRKStockPosition', ['$resource', function($resource){
    return $resource('/api/rrk/stock-position-for-date/:id',{},{
    } );
  }]);

bomServices.factory('RRKStockPristineDate', ['$resource', function($resource){
    return $resource('/api/rrk/stock-pristine-date',{},{
    } );
  }]);
