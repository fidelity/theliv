/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { KubePlatformComponent } from './kube-platform.component';
import { KubernetesService } from '../services/kubernetes.service';
import { RouterTestingModule } from '@angular/router/testing';
import { ActivatedRoute, Router } from '@angular/router';
import { BehaviorSubject, Observable, of, throwError } from 'rxjs';

class MockedKubernetesService {
  public resourceList$: BehaviorSubject<any> = new BehaviorSubject<any>([]);
  public selectedClusters$: BehaviorSubject<any> = new BehaviorSubject<any>('');
  public selectedNs$: BehaviorSubject<any> = new BehaviorSubject<any>('');

  getAllNamespaces(): Observable<any> {
      // return of({
      //   items: [{name: 'ns1'}, {name: 'ns2'}]
      // });
      return of([
        'ns1', 'ns2'
      ]);
  }
  getResourceInfo(): Observable<any> {
    return of([]);
  }
  getClusters(): Observable<any> {
    return of(['cluster-1', 'cluster-2']);
  }
  getNSByCluster(): Observable<any> {
    return of(['ns-1', 'ns-2']);
  }
  getDetects(c: string, ns: string): Observable<any> {
    return of([{
      name: 'resourcegroup-1'
    },
    {
      name: 'resourcegroup-2'
    }]);
  }
}

let MockactiveRoute: any;

describe('KubePlatformComponent', () => {
  let component: KubePlatformComponent;
  let fixture: ComponentFixture<KubePlatformComponent>;
  let mockRouter: Router;
  let kubernetesService: KubernetesService;
  MockactiveRoute = {
    queryParamMap: of(
      {
          params: {cluster: 'cluster', namespace: 'ns'},
          get: (key: any) => {
              if (key === 'cluster') {
                  return 'cluster';
              } else if (key === 'namespace') {
                  return 'ns';
              } else {
                  return '';
              }
          }
      }
      )
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [
        RouterTestingModule
      ],
      declarations: [ KubePlatformComponent ],
      providers: [
        {
            provide: KubernetesService,
            useClass: MockedKubernetesService
        },
        {
            provide: ActivatedRoute,
            useValue: MockactiveRoute
        }
      ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    kubernetesService = TestBed.inject(KubernetesService);
    fixture = TestBed.createComponent(KubePlatformComponent);
    component = fixture.componentInstance;
    mockRouter = TestBed.inject(Router);
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  // it('should call changeSelection event when click check with "All" value', () => {
  //   const el = fixture.nativeElement.querySelector('.resource-style');
  //   el.value = 'All';
  //   el.checked = true;
  //   el.dispatchEvent(new Event('change'));
  //   expect(component.resourceGroups.length).toEqual(2);
  // });

  // it('should call changeSelection event when click check with "Pods" value', () => {
  //   const el = fixture.nativeElement.querySelector('.resource-style');
  //   el.value = 'Pods';
  //   el.checked = true;
  //   el.dispatchEvent(new Event('change'));
  //   expect(component.type.includes('Pods')).toBeTruthy();
  // });

  // it('should call changeSelection event when click check with "Pods" value and error response', () => {
  //   kubernetesService.getDetects = jasmine.createSpy().and.returnValue(throwError({err: {statusText: 'test'}}));
  //   const el = fixture.nativeElement.querySelector('.resource-style');
  //   el.value = 'Pods';
  //   el.checked = true;
  //   el.dispatchEvent(new Event('change'));
  //   expect(component.type.includes('Pods')).toBeTruthy();
  // });

  // it('should call changeSelection event when click uncheck with "Pods" value', () => {
  //   const el = fixture.nativeElement.querySelector('.resource-style');
  //   component.type = ['Pods'];
  //   el.value = 'Pods';
  //   el.checked = false;
  //   el.dispatchEvent(new Event('change'));
  //   expect(component.type.includes('Pods')).toBeFalsy();
  // });

  it('should get namespace by cluster name', () => {
    const navigateSpy = spyOn(mockRouter, 'navigate');
    component.selectedClusters = 'cluster1';
    component.getNSByCluster();
    expect(navigateSpy).toHaveBeenCalled();
    expect(component.namespaces.length).toEqual(2);
  });

  it('should get namespace by cluster name with error response', () => {
    kubernetesService.getAllNamespaces = jasmine.createSpy().and.returnValue(throwError({err: {statusText: 'test'}}));
    component.selectedClusters = 'cluster1';
    component.getNSByCluster();
    expect(component.namespaces.length).toEqual(0);
  });

  it('should call getSelectedQuery event when click ns selector', () => {
    const navigateSpy = spyOn(mockRouter, 'navigate');
    component.selectedClusters = 'cluster1';
    const el = fixture.nativeElement.querySelector('select[name="namespace"]');
    el.value = 'ns1';
    el.dispatchEvent(new Event('change'));
    expect(navigateSpy).toHaveBeenCalled();
    // expect(navigateSpy).toHaveBeenCalledWith([`/kubernetes/cluster1/ns1`]);
  });
});

