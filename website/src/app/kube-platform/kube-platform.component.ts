import { Component, OnInit } from '@angular/core';
import { Observable, of } from 'rxjs';
import { ActivatedRoute, Router } from '@angular/router';
import { faSearch, faTimes, faSpinner, faBell, faCheck} from '@fortawesome/free-solid-svg-icons';
import { KubernetesService } from '../services/kubernetes.service';
import { debounceTime, delay, distinctUntilChanged, map, startWith, switchMap } from 'rxjs/operators';
import { FormControl } from '@angular/forms';

export interface NamespaceOption {
  text: string
  value: string
}

@Component({
  selector: 'app-kube-platform',
  templateUrl: './kube-platform.component.html',
  styleUrls: ['./kube-platform.component.scss']
})
export class KubePlatformComponent implements OnInit {
  faSearch = faSearch;
  faTimes = faTimes;
  faSpinner = faSpinner;
  faBell = faBell;
  faCheck = faCheck;
  loading = false;
  public resourceTypes: any;
  public proDomains: any;
  public senLevels: any;
  public resourceStatus: any;
  resourceGroups: any;
  allResources: any;
  isGroupContent = false;
  count = -1;
  sortBy = 'time';
  now = new Date();

  type: any;
  selectedNs = '';
  selectedClusters = '';
  //namespaces: any;
  clusters: any;

  clusterFormControl = new FormControl();
  clusterOptions: Observable<string[]> | undefined;
  namespaces: NamespaceOption[] = []

  constructor(
    private kubeService: KubernetesService,
    private route: ActivatedRoute,
    private router: Router) { }

  ngOnInit(): void {
    this.resourceGroups = null;
    // this.router.routeReuseStrategy.shouldReuseRoute = () => false;h
    this.kubeService.getClusters().subscribe((res: any) => {
      if (res) {
        this.clusters = res;
        this.clusterOptions = this.clusterFormControl.valueChanges.pipe(
          startWith(''),
          map(value => {
            const filterValue = value.toLowerCase();
            return this.clusters.filter((option: string) =>  
              option.toLowerCase().indexOf(filterValue) >= 0 
            );
          })
        );
      }
    }, (err: any) => {
      console.log('Get Clusters Information Error: ', err);
    });

    this.route.queryParamMap.subscribe((params: any) => {
      this.resourceGroups = [];
      if (params.get('cluster')) {
        this.selectedClusters = params.get('cluster')
      }
      this.kubeService.selectedClusters$.next(this.selectedClusters);

      if (params.get('namespace')) {
        this.selectedNs = params.get('namespace');
      }
      this.kubeService.selectedNs$.next(this.selectedNs);

      this.getNamespaces()
      if (this.selectedClusters && this.selectedNs) {
        this.getKubeResourceInfo();
      }
    });
  }

  getKubeResourceInfo(): void {
    this.loading = true;
    this.kubeService.getDetects(this.selectedClusters, this.selectedNs).subscribe(
      (res: any) => {
        if (res) {
          this.resourceGroups = res;
          this.allResources = res;
          this.kubeService.resourceList$.next(this.resourceGroups);
          this.isGroupContent = true;
          this.count = res.length;
          this.loading = false;

          this.getResourceTypeFilter()
          this.getProblemDomainFilter()
        }
      },
      (err: any) => {
        console.log('Get Kube Information Error: ', err);
      }
    );
  }

  getResourceTypeFilter(): void {
    var list: any[] = [];
    this.allResources.forEach((r: any) => {
      if (!list.find((item: any) => item.name.includes(r.topResourceType))) {
        var n = r.topResourceType;
        var c = 1;
        var resource = {name: n, count: c}
        list.push(resource)
      } else{
        var obj = list.find((item: any) => item.name === r.topResourceType)
        obj.count = obj.count + 1;
      }
    });
    var totalCount = 0;
    list.forEach((item: any)=>{
      totalCount = totalCount+ item.count;
    })
    list.push({name: 'All Types', count: totalCount})
    console.log(list);
    this.resourceTypes = list;
  }

  getProblemDomainFilter(): void {
    var list: any[] = [];
    this.allResources.forEach((r: any) => {
      if (!list.find((item: any) => item.name.includes(r.rootCause.name))) {
        var n = r.rootCause.name;
        var c = 1;
        var resource = {name: n, count: c}
        list.push(resource)
      } else{
        var obj = list.find((item: any) => item.name === r.rootCause.name)
        obj.count = obj.count + 1;
      }
    });
    var totalCount = 0;
    list.forEach((item: any)=>{
      totalCount = totalCount+ item.count;
    })
    list.push({name: 'All Domains', count: totalCount })
    console.log(list);
    this.proDomains = list;
  }


  filter: any[] = [];
  changeSelection(e: any, cat: string): void {
    if (e.target.checked && e.target.value) {
      console.log('checked: ', e.target.value);
      if (e.target.value!=="All Domains" && e.target.value!=="All Types" ) {
        this.filter.push(e.target.value);
        switch (cat) {
          case 'type':
            this.filterRescourceByType(this.filter);
            break;
          case 'domain':
            this.filterRescourceByDomain(this.filter);
            break;
          default:
            break;
        }
      } else {
        this.resourceGroups = this.allResources;
        this.count = this.resourceGroups.length;
      }
    } else if (!e.target.checked && e.target.value) {
      this.filter.splice(this.filter.findIndex((element) => element == e.target.value), 1);
      console.log('unchecked: ', e.target.value);
      if (this.filter.length>0){
        switch (cat) {
          case 'type':
              this.filterRescourceByType(this.filter);
            break;
          case 'domain':
            this.filterRescourceByDomain(this.filter);
            break;
          default:
            break;
          }
      } else{
        this.resourceGroups = this.allResources;
        this.count = this.resourceGroups.length;
      }
    }
    this.kubeService.resourceList$.next(this.resourceGroups);
    // else if (this.type.indexOf(e.target.value) > -1) {
    //   const index = this.type.indexOf(e.target.value, 0);
    //   this.type.splice(index, 1);
    // }
  }


  filterRescourceByDomain(filter: string[]): void {
    var resultList: any;
    resultList=this.allResources.filter((r: any) => filter.includes(r.rootCause.name));
    this.resourceGroups = resultList;
  }

  filterRescourceByType(filter: string[]): void {
    var resultList: any;
    resultList=this.allResources.filter((r: any) => filter.includes(r.topResourceType));
    this.resourceGroups = resultList;
  }

  checkClusterBlank(){
    if (this.selectedClusters == '') {
      this.selectedNs = ''
      this.namespaces = []
      this.resourceGroups = []
      this.resourceTypes = []
      this.proDomains = []
    }
  }

  getNSByCluster(): void {
    this.router.navigate(['kubernetes']);
    this.resourceGroups = [];
    this.selectedNs = '';
    this.kubeService.selectedNs$.next(this.selectedNs);
    this.kubeService.resourceList$.next(this.resourceGroups);
    this.getNamespaces()
  }

  getNamespaces(): void {
    if (this.selectedClusters) {
      this.kubeService.getAllNamespaces(this.selectedClusters).subscribe(
        (res: any) => {
          if (res) {
            this.namespaces = []
            res.forEach((ele: string) => {
              this.namespaces.push({text: ele, value: ele})
            });
          }
        },
        (err: any) => {
          console.log(`Get Namespace in Cluster ${this.selectedClusters} Error: `, err);
        }
      );
    }
  }

  getSelectedQuery(e: any): void {
    this.router.navigate(['kubernetes']);
    this.resourceGroups = [];
    this.kubeService.selectedClusters$.next(this.selectedClusters);
    this.kubeService.selectedNs$.next(this.selectedNs);
    if (this.selectedClusters && this.selectedNs) {
      this.getKubeResourceInfo();
    }
  }

}
