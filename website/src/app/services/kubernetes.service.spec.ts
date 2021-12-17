import { KubernetesService } from './kubernetes.service';
import { HttpClient } from '@angular/common/http';
import { of, throwError, BehaviorSubject } from 'rxjs';

describe('KubernetesService', () => {
  let service: KubernetesService;
  let mockHttpService: HttpClient;

  beforeEach(() => {
    mockHttpService = jasmine.createSpyObj('HttpClient', ['get']);
    service = new KubernetesService(mockHttpService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should call getClusters method', () => {
    (mockHttpService.get as jasmine.Spy).and.returnValue(of(['cluster-1', 'cluster-2']));
    service.getClusters();
    expect(mockHttpService.get).toHaveBeenCalled();
    expect(mockHttpService.get).toHaveBeenCalledWith('https://theliv-dev.us-east-1.eccdevpe6.aws-nonprod.fmrcloud.com/theliv-api/v1/clusters');
  });

  it('should call getAllNamespaces methods', () => {
    (mockHttpService.get as jasmine.Spy).and.returnValue(of({name: 'getProjectTest'}));
    service.getAllNamespaces('clustername');
    expect(mockHttpService.get).toHaveBeenCalled();
    expect(mockHttpService.get).toHaveBeenCalledWith('https://theliv-dev.us-east-1.eccdevpe6.aws-nonprod.fmrcloud.com/theliv-api/v1/clusters/clustername/namespaces');
  });

  it('should throw error if error is reponded when call getAllNamespaces', () => {
    (mockHttpService.get as jasmine.Spy).and.returnValue(throwError({}));
    service.getAllNamespaces('').subscribe(res => {},
    err => {
      expect(mockHttpService.get).toHaveBeenCalled();
      expect(err).toBeDefined();
    });
  });

  it('should call getAllNamespaces methods with null error', () => {
    (mockHttpService.get as jasmine.Spy).and.returnValue(throwError(null));
    service.getAllNamespaces('cluster').subscribe(res => {},
    err => {
      expect(mockHttpService.get).toHaveBeenCalled();
      expect(err).toEqual('Service error');
    });
  });

  it('should call getResourceInfo methods', () => {
    (mockHttpService.get as jasmine.Spy).and.returnValue(of({name: 'getProjectTest'}));
    service.getDetects('cn', 'ns');
    expect(mockHttpService.get).toHaveBeenCalled();
    expect(mockHttpService.get).toHaveBeenCalledWith('https://theliv-dev.us-east-1.eccdevpe6.aws-nonprod.fmrcloud.com/theliv-api/v1/detector/cn/ns/detect');
  });

  it('should throw error if error is reponded when call getResourceInfo', () => {
    (mockHttpService.get as jasmine.Spy).and.returnValue(throwError({}));
    service.getDetects('', '').subscribe(res => {},
    err => {
      expect(mockHttpService.get).toHaveBeenCalled();
      expect(err).toBeDefined();
    });
  });

});
