/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
import { Component, OnInit } from '@angular/core';
import { faSearch, faTimes, faSpinner, faBell, faCheck, faPencilAlt, faAngleDown, faAngleUp, faVideo} from '@fortawesome/free-solid-svg-icons';
import { Observable } from 'rxjs';
import { ActivatedRoute, Router } from '@angular/router';
import { KubernetesService } from '../../services/kubernetes.service';
import { map, startWith } from 'rxjs/operators';
import { FormControl } from '@angular/forms';
import { DomSanitizer } from '@angular/platform-browser';

export interface NamespaceOption {
  text: string
  value: string
}

@Component({
  selector: 'app-home-page',
  templateUrl: './home-page.component.html',
  styleUrls: ['./home-page.component.scss']
})
export class HomePageComponent implements OnInit {
  faSearch = faSearch;
  faTimes = faTimes;
  faSpinner = faSpinner;
  faBell = faBell;
  faCheck = faCheck;
  faPencialAlt = faPencilAlt;
  faAngleDown = faAngleDown;
  faAngleUp = faAngleUp;
  faVideo = faVideo;
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
  clusters: any;

  clusterFormControl = new FormControl();
  clusterOptions: Observable<string[]> | undefined;
  namespaces: NamespaceOption[] = []
  clusterInputing = false

  events: any;
  sortedData: any;
  isAsc=false;
  active='time';
  feedback = '';
  gridToggle = false;
  configInfo: any;
  urlSafe: any;

  constructor(    
    private kubeService: KubernetesService,
    private router: Router,
    public sanitizer: DomSanitizer) { }

    ngOnInit(): void {
      this.resourceGroups = null;
      this.loading = true;
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

      this.kubeService.getConfigInfo().subscribe((res: any) => {
        if (res) {
          this.configInfo = res;
          this.urlSafe = this.sanitizer.bypassSecurityTrustResourceUrl(this.configInfo.videoLink);
          this.loading = false;
        }
      }, (err: any) => {
        console.log('Get Config Information Error: ', err);
      });

    }

  
    checkClusterBlank(){
      this.clusterInputing = true
      if (this.selectedClusters == '') {
        this.selectedNs = ''
        this.namespaces = []
        this.resourceGroups = []
        this.resourceTypes = []
        this.proDomains = []
        this.events = []
      }
    }
  
    getNSByCluster(): void {
      this.clusterInputing = false;
      this.getNamespaces();
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
      this.router.navigate(['kubernetes'], { queryParams: { cluster: this.selectedClusters, namespace: e.value } });
    }

  }
